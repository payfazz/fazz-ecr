package oidctoken

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/payfazz/go-errors/v2"
	"github.com/payfazz/go-handler/v2"
	"github.com/payfazz/go-handler/v2/defresponse"

	oidcconfig "github.com/payfazz/fazz-ecr/config/oidc"
	"github.com/payfazz/fazz-ecr/util/randstring"
)

func GetToken(callback func(string) (string, error)) error {
	// this funciton is not thread-safe, do not call it from multiple go routine at the same time

	if token := os.Getenv("FAZZ_ECR_TOKEN"); token != "" {
		_, err := callback(token)
		return err
	}

	if os.Getenv("CI") != "" {
		return errors.New("empty FAZZ_ECR_TOKEN enviroment variable is not supported in CI environment")
	}

	cache := loadTokenCache()
	provider := loadProviderCache()

	if v, ok := cache[oidcconfig.Issuer]; ok {
		if time.Now().Unix() < v.Exp {
			_, err := callback(v.IDToken)
			return err
		}
		if refresh := v.RefreshToken; refresh != "" {
			if cont, err := func() (bool, error) {
				token, refresh, err := provider.refreshIDToken(oidcconfig.Issuer, oidcconfig.ClientID, refresh)
				if err != nil {
					return true, err
				}

				exp, err := getTokenExp(token)
				if err != nil {
					return true, err
				}

				cache[oidcconfig.Issuer] = tokenCacheItem{
					IDToken:      token,
					RefreshToken: refresh,
					Exp:          exp,
				}
				cache.save()

				_, err = callback(token)
				return false, err
			}(); !cont {
				return err
			}
		}
	}

	redirect := fmt.Sprintf("http://localhost:%d", oidcconfig.CallbackPort)
	state := randstring.Get(16)

	auth, err := provider.getAuthUri(oidcconfig.Issuer, oidcconfig.ClientID, redirect, state)
	if err != nil {
		return err
	}

	handled := uint32(0)

	type resp struct {
		text   string
		status int
	}

	loginHitCh := make(chan struct{}, 1)
	codeCh := make(chan string, 1)
	respCh := make(chan resp, 1)

	server := http.Server{
		Addr: strings.TrimPrefix(redirect, "http://"),
		Handler: handler.Of(func(r *http.Request) http.HandlerFunc {
			path := r.URL.EscapedPath()
			if path == "/login" {
				select {
				case loginHitCh <- struct{}{}:
				default:
				}
				return defresponse.Redirect(302, auth)
			} else if path != "/" {
				return defresponse.Status(404)
			}

			if r.URL.Query().Get("state") != state {
				return defresponse.Text(400, `invalid "state"`)
			}

			code := r.URL.Query().Get("code")
			if code == "" {
				return defresponse.Text(400, `"code" is empty`)
			}

			if !atomic.CompareAndSwapUint32(&handled, 0, 1) {
				return defresponse.Text(400, "cannot call callback multiple time")
			}

			codeCh <- code
			resp := <-respCh

			return defresponse.Text(resp.status, resp.text)
		}),
	}

	serverErrCh := make(chan error, 1)
	errors.Go(
		func(err error) { serverErrCh <- err },
		func() error { return errors.Trace(server.ListenAndServe()) },
	)
	defer server.Shutdown(context.Background())

	// this is to make sure that server is running first
	select {
	case err := <-serverErrCh:
		return err
	case <-time.After(500 * time.Millisecond):
	}

	loginLink := fmt.Sprintf("http://%s/login", server.Addr)
	if err := openBrowser(loginLink); err != nil {
		return errors.Trace(err)
	}

	select {
	case err := <-serverErrCh:
		return err
	case <-time.After(10 * time.Second):
		return errors.Errorf("%s is not opened after 10 second", loginLink)
	case <-loginHitCh:
	}

	processCode := func(code string) (string, error) {
		token, refresh, err := provider.getIDToken(oidcconfig.Issuer, oidcconfig.ClientID, redirect, code)
		if err != nil {
			return "", err
		}

		exp, err := getTokenExp(token)
		if err != nil {
			return "", err
		}

		cache[oidcconfig.Issuer] = tokenCacheItem{
			IDToken:      token,
			RefreshToken: refresh,
			Exp:          exp,
		}
		cache.save()

		res, err := callback(token)
		if err != nil {
			return "", err
		}

		return res, nil
	}

	select {
	case err := <-serverErrCh:
		return err
	case <-time.After(5 * time.Minute):
		return errors.Errorf("timed out after waiting for 5 minutes")
	case code := <-codeCh:
		text, err := processCode(code)
		if err != nil {
			respCh <- resp{text: err.Error(), status: 500}
		} else {
			respCh <- resp{text: text, status: 200}
		}
		return err
	}
}

func getTokenExp(token string) (int64, error) {
	tokenParts := strings.Split(token, ".")
	if len(tokenParts) < 2 {
		return 0, errors.Errorf("invalid token from oidc")
	}

	tokenBodyRaw, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return 0, errors.Trace(err)
	}

	var tokenBody struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(tokenBodyRaw, &tokenBody); err != nil {
		return 0, errors.Trace(err)
	}

	return tokenBody.Exp - 10, nil
}
