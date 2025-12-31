# GitLab Webhook Example

This example sends a GitLab merge request webhook to the local githooks server.

## Prerequisites
- githooks running on `http://localhost:8080`
- `GITLAB_WEBHOOK_SECRET` set to the same value in your config (optional)
- Optional SCM auth: `GITLAB_ACCESS_TOKEN`

## Run
```sh
export GITLAB_WEBHOOK_SECRET=devsecret
export GITLAB_ACCESS_TOKEN=glpat-xxxx
./scripts/send_webhook.sh gitlab "Merge Request Hook" example/gitlab/merge_request.json
```

Push event:
```sh
./scripts/send_webhook.sh gitlab "Push Hook" example/gitlab/push.json
```

Example config (rules) is in `example/gitlab/app.yaml`.
