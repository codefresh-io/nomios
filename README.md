# Nomios - Codefresh DockerHub Event Provider

[![Codefresh build status](https://g.codefresh.io/api/badges/build?repoOwner=codefresh-io&repoName=nomios&branch=master&pipelineName=nomios&accountName=codefresh-inc&type=cf-1)](https://g.codefresh.io/repositories/codefresh-io/nomios/builds?filter=trigger:build;branch:master;service:5a2f9ac17e524c00017a5970~nomios) [![Go Report Card](https://goreportcard.com/badge/github.com/codefresh-io/nomios)](https://goreportcard.com/report/github.com/codefresh-io/nomios) [![codecov](https://codecov.io/gh/codefresh-io/nomios/branch/master/graph/badge.svg)](https://codecov.io/gh/codefresh-io/nomios)

[![](https://images.microbadger.com/badges/image/codefresh/nomios.svg)](http://microbadger.com/images/codefresh/nomios) [![](https://images.microbadger.com/badges/commit/codefresh/nomios.svg)](https://microbadger.com/images/codefresh/nomios) [![Docker badge](https://img.shields.io/docker/pulls/codefresh/nomios.svg)](https://hub.docker.com/r/codefresh/nomios/)

Codefresh *DockerHub Event Provider*, code named *Nomios* (son of *Hermes*) notifies [Hermes](https://github.com/codefresh-io/hermes) service when a new image pushed to a DockerHub.

*Nomios* is a DockerHub webhook server. If properly configured (see bellow), it will receive an event for every `docker push` command. *Nomios* understand DockerHub webhook payload and generates "normalized* event that it sends to *Hermes* trigger manager.

## Normalized event

POST ${HERMES_SERVICE}/trigger/${event}

```json
{
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

### Fields

- URL: `event` - event URI in form `registry:dockerhub:<namespace>:<name>:push`
- PAYLOAD: `secret` - webhook secret
- PAYLOAD: `original` - original DockerHub `push` event JSON payload
- PAYLOAD: `variables` - set of variables, extracted from the event payload: `namespace`, `name`, `tag`, `pusher`, `pushed_at`

## Configure DockerHub webhook

Configuring webhooks for DockerHub, requires manual work.

To configure webhooks, visit `https://hub.docker.com/r/<USERNAME>/<REPOSITORY>/~/settings/webhooks/`.
You can get more information, reading the official Docker [documentation](https://docs.docker.com/docker-hub/webhooks/)

### DockerHub webhook security

DockerHub webhook has no built-in security mechanism. Codefresh adds basic security to avoid webhook abuse.

When adding a new trigger into *Hermes* trigger manager server, specify some secret (`MYSECRET1234` for example). Use different secret for different DockerHub event URIs. Use selected secret as `secret` parameter in webhook URL. For example `https://g.codefresh.io/dockerhub?secret=MYSECRET1234`.

*Nomios* will extract this secret from URL and will pass it to *Hermes* service for validation. If the secret hs no match, *Hermes* will not trigger Codefresh pipeline execution.

## Running Nomios service

Run the `nomios server` command to start *Nomios* DockerHub event provider.

```sh
NAME:
   nomios server - start Nomios DockerHub webhook handler server

USAGE:
   nomios server [command options] [arguments...]

DESCRIPTION:
   Run DockerHub WebHook handler server. Process and send normalized event payload to the Codefresh Hermes trigger manager service to invoke associated Codefresh pipelines.

OPTIONS:
   --hermes value, --hm value  Codefresh Hermes service (default: "http://hermes/") [$HERMES_SERVICE]
   --token value, -t value     Codefresh Hermes API token (default: "TOKEN") [$HERMES_TOKEN]
```

## Building Nomios

`nomios` requires Go SDK to build.

1. Clone this repository into `$GOPATH/src/github.com/codefresh-io/nomios`
1. Run `hack/build.sh` helper script or `go build cmd/main.go`%
1. Run `hack/test.sh` to run all tests

