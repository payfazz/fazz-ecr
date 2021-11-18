# fazz-ecr

Tools for interacting with the ECR Docker registry
`322727087874.dkr.ecr.ap-southeast-1.amazonaws.com`.

This repo provide two utilities:

- `docker-credential-fazz-ecr`: Credential helper for Docker client.
- `fazz-ecr-create-repo`: Helper to create new repository on the
  `322727087874.dkr.ecr.ap-southeast-1.amazonaws.com` registry.

User permissions to repositories are determined by membership of Google groups.

## How to install

Run `go install github.com/payfazz/fazz-ecr/cmd/...@latest` to install both
utilities in your `$GOBIN`. Put `$GOBIN` in your environment to make the program accessible from your terminal.

## Quickstart
1. Make sure that docker is installed
   ```
   which docker
   ```
   If docker is not found, install docker first by following the instructions [here]("https://docs.docker.com/engine/install/")


2. Check for the file in `{HOME}/.docker/config.json`. If it is not available create the file with `{}` (empty json object). Or run this command:
   
   ```sh
   [ ! -f ${HOME}/.docker/config.json ] && mkdir -p ${HOME}/.docker && echo "{}" > ${HOME}/.docker/config.json
   ```
3. To generate fazz-ecr auth config for docker, run the following on terminal:

   ```
   docker-credential-fazz-ecr update-config
   ```
4. You can now create a repository based on your account. For example, you are creating a service named `foo` and have built the docker image. Run the following command to create the repo
   ```sh
   fazz-ecr-create-repo 322727087874.dkr.ecr.ap-southeast-1.amazonaws.com/your_name@fazzfinancial.com/foo
   ```
   After creating the repository, you can now push the docker image using the ususal docker command
   ```
   docker push 322727087874.dkr.ecr.ap-southeast-1.amazonaws.com/your_name@fazzfinancial.com/foo:1.0.0
   ```
   ---

   **NOTE**

   This repository objects is immutable, which means that an image with the same tag cannot be pushed twice. 

   ---
   The same rule is applicable to repository with your team's name. For example you are in `payfazz` team, you have access to create a repository on 
   `322727087874.dkr.ecr.ap-southeast-1.amazonaws.com/payfazz/*`.

   e.g. you can push on this repository if you are registered at payfazz team
   ```bash
   docker push 322727087874.dkr.ecr.ap-southeast-1.amazonaws.com/payfazz/authfazz:v1.0.0
   ```

   To find which team you are registered at, please ask your team lead or any of the SRE team.


## How to use in GitHub Actions

Use `payfazz/setup-fazz-ecr-action@v1` action in your workflow file. Because CI
environment is not interactive, `FAZZ_ECR_TOKEN` environment variable must be
set.

# `docker-credential-fazz-ecr`

`docker-credential-fazz-ecr` provides sub-command:

- `update-config`
- `login`
- `list-access`
- `get`

`docker-credential-fazz-ecr update-config` is used to update
`~/.docker/config.json` so that any access to registry
`322727087874.dkr.ecr.ap-southeast-1.amazonaws.com` is using `fazz-ecr`
credential

`docker-credential-fazz-ecr login` is used to remove old credential and retrieve
new credential, this command will open web browser to authenticate your identity

`docker-credential-fazz-ecr list-access` is used to print which repository is
accessible by current credentials

`docker-credential-fazz-ecr get` is used internally by `docker` command line
