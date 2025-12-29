# Boilerplate Worker

This folder provides a starting point for building a worker that consumes Githooks topics.

## Run

1. Start the server:
```sh
export GITHUB_WEBHOOK_SECRET=devsecret

go run ./main.go -config boilerplate/worker/config.yaml
```

2. Start the worker:
```sh
go run ./boilerplate/worker -config boilerplate/worker/config.yaml
```

3. Send a webhook (example):
```sh
./scripts/send_webhook.sh github pull_request example/github/pull_request.json
```

## Customize
- Update `boilerplate/worker/controllers/` with your handlers.
- Update `boilerplate/worker/main.go` to register handlers.
- Update `boilerplate/worker/config.yaml` with your broker config and rules.

## Env
Copy the env file and update secrets:
```sh
cp boilerplate/worker/.env.example .env
```

## Makefile
Common commands:
```sh
make deps-up
make run-server
make run-worker
```

Notes:
- Run `make` from `boilerplate/worker/`.
- Override paths if you copied the boilerplate elsewhere (e.g., `make ROOT=. run-worker`).

## Local Dependencies
Start RabbitMQ with compose:
```sh
docker compose -f boilerplate/worker/docker-compose.yaml up -d
```

## Docker
Build and run the worker container:
```sh
docker build -f boilerplate/worker/Dockerfile -t githooks-worker .
docker run --rm -e GITHUB_WEBHOOK_SECRET=devsecret githooks-worker -config /app/config.yaml
```

## Helm
You can deploy a worker using the Helm chart:

```sh
helm install githooks-worker ./charts/githooks-worker \
  --set image.repository=ghcr.io/your-org/your-worker \
  --set image.tag=latest \
  --set-file configYaml=boilerplate/worker/config.yaml
```
