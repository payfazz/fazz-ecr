package main

import (
	"time"

	"github.com/payfazz/fazz-ecr/config"
	"github.com/payfazz/fazz-ecr/pkg/types"
	"github.com/payfazz/fazz-ecr/util/jsonfile"
)

type cache map[string]types.Cred

func loadCache() cache {
	var ret cache
	if err := jsonfile.Read(config.CacheFileDockerCreds, &ret); err != nil {
		return make(cache)
	}

	var expList []string
	now := time.Now().Unix()
	for k, v := range ret {
		if v.Exp <= now {
			expList = append(expList, k)
		}
	}
	for _, k := range expList {
		delete(ret, k)
	}

	return ret
}

func (c cache) save() {
	jsonfile.Write(config.CacheFileDockerCreds, c)
}