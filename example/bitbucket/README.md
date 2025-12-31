# Bitbucket Webhook Example

This example sends a Bitbucket pull request webhook to the local githooks server.

## Prerequisites
- githooks running on `http://localhost:8080`
- `BITBUCKET_WEBHOOK_SECRET` set to the same UUID in your config (optional)
- Optional SCM auth: `BITBUCKET_ACCESS_TOKEN`

## Run
```sh
export BITBUCKET_WEBHOOK_SECRET=example-uuid
export BITBUCKET_ACCESS_TOKEN=bb-xxxx
./scripts/send_webhook.sh bitbucket pullrequest:created example/bitbucket/pull_request.json
```

Repo push event:
```sh
./scripts/send_webhook.sh bitbucket repo:push example/bitbucket/repo_push.json
```

Example config (rules) is in `example/bitbucket/app.yaml`.
