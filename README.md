# fazz-ecr

Tools for interacting with the ECR Docker registry `322727087874.dkr.ecr.ap-southeast-1.amazonaws.com`.

This repo provide two utilities:

- `docker-credential-fazz-ecr`: Credential helper for Docker client.
- `fazz-ecr-create-repo`: Helper to create new repository on the `322727087874.dkr.ecr.ap-southeast-1.amazonaws.com` registry.

User permissions to repositories are determined by the membership of Google groups.


## How to install

Run `go install github.com/payfazz/fazz-ecr/cmd/...@latest` to install both utilities in your `$GOBIN`. Put `$GOBIN` in your `$PATH` to make it easier to access the programs.


## Quickstart

1. Make sure that docker is installed
   ```sh
   which docker
   ```
   If docker is not found, install docker first by following the instructions [here]("https://docs.docker.com/engine/install/")

2. Check for the file in `{HOME}/.docker/config.json`. If it is not available create the file with `{}` (empty json object). Or run this command:
   ```sh
   [ ! -f ${HOME}/.docker/config.json ] && mkdir -p ${HOME}/.docker && echo "{}" > ${HOME}/.docker/config.json
   ```

3. To generate fazz-ecr auth config for docker, run the following on terminal:
   ```sh
   docker-credential-fazz-ecr update-config
   ```

4. You can now create a repository based on your account. For example, if your email is `myname@fazzfinancial.com`, and you want to create a container image repository for service named `myservice`, run the following command:
   ```sh
   fazz-ecr-create-repo 322727087874.dkr.ecr.ap-southeast-1.amazonaws.com/myname-fazzfinancial-com/myservice
   ```

   The same rule is applicable to repository with your team's name. For example, if you are a member of `auth-payfazz@fazzfinancial.com`, you can create repositories on `322727087874.dkr.ecr.ap-southeast-1.amazonaws.com/payfazz/*`.

   You cannot push image with the same tag more than once (Every image:tag is immutable).


## How to use in GitHub Actions

Use `payfazz/setup-fazz-ecr-action@v1` action in your workflow file. `FAZZ_ECR_TOKEN` environment variable must be set.


## `docker-credential-fazz-ecr`

`docker-credential-fazz-ecr` provides the following sub-commands:
- `update-config`
- `login`
- `list-access`
- `get`

`docker-credential-fazz-ecr update-config` is used to update `~/.docker/config.json` to configure authentication for `322727087874.dkr.ecr.ap-southeast-1.amazonaws.com` repository.

`docker-credential-fazz-ecr login` is used to remove old credential and retrieve new credential. This command will open your web browser to authenticate your identity.

`docker-credential-fazz-ecr list-access` is used to list all repositories which is accessible by your credential.

`docker-credential-fazz-ecr get` is used internally by `docker` command line


## Lambda

Both client utilities will do an HTTP call to an AWS Lambda which is proxied by AWS API Gateway. This lambda handles the actual authentication and authorization logic. The lambda code is located on `aws-lambda/fazz-ecr` directory, and you can deploy the code with the following commands:
```sh
# Make sure you are in the aws-lambda/fazz-ecr directory
CGO_ENABLED=0 GOOS=linux go build .
zip function.zip fazz-ecr
aws lambda update-function-code --function-name fazz-ecr --region ap-southeast-1 --zip-file fileb://function.zip
rm function.zip fazz-ecr
```
