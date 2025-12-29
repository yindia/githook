# Boilerplate Worker

This folder provides a starting point for building a worker that consumes Githooks topics.

## Run

1. Start the server:
```sh
export GITHUB_WEBHOOK_SECRET=devsecret

go run ./main.go -config boilerplate/config.yaml
```

2. Start the worker:
```sh
go run ./boilerplate/worker
```

3. Send a webhook (example):
```sh
./example/github/send_webhook.sh
```

## Customize
- Update `boilerplate/worker/main.go` handlers and topics.
- Update `boilerplate/config.yaml` with your broker config and rules.
