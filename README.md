# Nomios - Codefresh DockerHub Event Provider

Codefresh *DockerHub Event Provider*, code named *Nomios* (son of *Hermes*) notifies [Hermes](https://github.com/codefresh-io/hermes) service when a new image pushed to a DockerHub.

*Nomios* is a DockerHub webhook server. If properly configured (see bellow), it will receive an event for every `docker push` command. *Nomios* understand DockerHub webhook payload and generates "normalized* event that it sends to *Hermes* trigger manager.

### Normalized event

```json
{
    "event": "index.docker.io:<namespace>:<name>:<tag>:push",
    "secret": "webhook secret",
    "original": "<original DockerHub webhook payload",
    "variables": {
        "namespace": "<image namespace>",
        "name": "<image name>",
        "tag": "<image tag>",
        "pusher": "<user that did a push command>",
        "pushed_at": "<RFC3339 formated timestamp>"
    }
}
```

#### Fields

- `event` - event URI in form `index.docker.io:<namespace>:<name>:<tag>:push`
- `secret` - webhook secret
- `original` - original DockerHub `push` event JSON payload
- `variables` - set of variables, extracted from the event payload: `namespace`, `name`, `tag`, `pusher`, `pushed_at`

## Configure DockerHub webhook

Configuring webhooks for DockerHub, requires manual work.

To configure webhooks, visit `https://hub.docker.com/r/<USERNAME>/<REPOSITORY>/~/settings/webhooks/`. 
You can get more information, reading the official Docker [documentation](https://docs.docker.com/docker-hub/webhooks/)

### DockerHub webhook security

DockerHub webhook has no built-in security mechanism. Codefresh adds basic security to avoid webhook abuse.

When adding a new trigger into *Hermes* trigger manager server, specify some secret (`MYSECRET1234` for example). Use different secret for different DockerHub event URIs. Use selected secret as `secret` parameter in webhook URL. For example `https://g.codefresh.io/dockerhub?secret=MYSECRET1234`.

*Nomios* will extract this secret from URL and will pass it to *Hermes* service for validation. If the secret hs no match, *Hermes* will not trigger Codefresh pipeline execution.

## Running Nomios service

Run the `dockerhub-provider server` command to start *Nomios* DockerHub event provider.

```sh
NAME:
   dockerhub-provider server - start dockerhub-provider webhook handler server

USAGE:
   dockerhub-provider server [command options] [arguments...]

DESCRIPTION:
   Run DockerHub WebHook handler server. Process and send normalized event payload to the Codefresh Hermes trigger manager service to invoke associated Codefresh pipelines.

OPTIONS:
   --hermes value, --hm value  Codefresh Hermes service (default: "http://hermes/") [$HERMES_SERVICE]
   --token value, -t value     Codefresh Hermes API token (default: "TOKEN") [$HERMES_TOKEN]
```

## Building Nomios

`nomios` requires Go SDK to build.

1. Clone this repository into `$GOPATH/src/github.com/codefresh-io/dockerhub-provider`
1. Run `hack/build.sh` helper script or `go build cmd/main.go`%
1. Run `hack/test.sh` to run all tests