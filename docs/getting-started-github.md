# Getting Started: GitHub

This guide walks through running the GitHub webhook example and creating a GitHub App for real payloads.

## 1) Prerequisites

- Go 1.21+
- Docker + Docker Compose
- A GitHub account

## 2) Start the local brokers

From the repo root:

```bash
docker compose up -d
```

This starts RabbitMQ, NATS Streaming, Kafka/Zookeeper, and Postgres for local development.

## 3) Run the server

Start the webhook server with the GitHub example config:

```bash
go run ./main.go -config example/github/app.yaml
```

You should see logs showing the server is listening on `:8080`.

## 4) Run the worker

In another terminal, run the GitHub worker example:

```bash
go run ./example/github/worker/main.go
```

The worker subscribes to the topics emitted by the GitHub example rules and logs matched events.

## 5) Send a test webhook

Use the bundled script to send a local GitHub pull_request event:

```bash
./scripts/send_webhook.sh github pull_request example/github/pull_request.json
```

You should see:

- The server log an event match and publish to a topic.
- The worker log the event handling.

## 6) Create a GitHub App (for real webhooks)

1. Open GitHub: **Settings** -> **Developer settings** -> **GitHub Apps** -> **New GitHub App**.
2. App name: `githooks-local` (or any name you like).
3. Homepage URL: `http://localhost:8080`.
4. Webhook URL: `http://localhost:8080/webhooks/github`.
5. Webhook secret: choose a random string. Save it.
6. Permissions:
   - Repository permissions: set **Pull requests** to **Read-only** (add more if needed).
7. Subscribe to events:
   - `Pull request`
   - `Push` (optional)
8. Create the app.

### Install the GitHub App

1. In the app settings, click **Install App**.
2. Choose a test repository and install.

### Update your config

Set the webhook secret in `example/github/app.yaml` or export it as an env var:

```yaml
providers:
  github:
    enabled: true
    path: /webhooks/github
    secret: ${GITHUB_WEBHOOK_SECRET}
```

Then:

```bash
export GITHUB_WEBHOOK_SECRET="your-secret"
go run ./main.go -config example/github/app.yaml
```

Now GitHub will send real webhook events to your local server.

## 7) Optional: use ngrok for remote webhooks

If GitHub needs to reach your machine from the internet:

```bash
ngrok http 8080
```

Update the GitHub App webhook URL to the ngrok URL.

## 8) Troubleshooting

- `missing X-Hub-Signature`: your webhook secret does not match.
- `no matching rules`: ensure rules in `example/github/app.yaml` match your payload.
- `connection refused`: make sure Docker Compose is running for broker drivers.
