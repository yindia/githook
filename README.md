# Githooks

Githooks is a config-driven webhook router for GitHub, GitLab, and Bitbucket. It normalizes inbound webhook events, evaluates them against YAML rules, and publishes matching events to Watermill topics for downstream consumers.

## Features
- Typed webhook parsing via go-playground/webhooks
- Provider-agnostic normalized event model
- JSONPath + boolean rule engine
- Watermill-backed publishing (gochannel, Kafka, NATS Streaming, AMQP, SQL, HTTP, RiverQueue)
- Stateless and horizontally scalable

## Architecture
```mermaid
flowchart LR
  A[Webhook Provider\nGitHub/GitLab/Bitbucket] --> B[go-playground/webhooks]
  B --> C[Provider Adapter]
  C --> D[Normalized Event]
  D --> E[Rule Engine]
  E --> F[Watermill Publisher]
  F --> G[AMQP/Kafka/NATS/SQL/HTTP/RiverQueue]
  F --> H[Workers/Consumers]
```

## Getting Started (Local)

1. Start dependencies:
```bash
docker compose up -d
```

2. Run the server:
```bash
export GITHUB_WEBHOOK_SECRET=devsecret

go run ./main.go -config app.docker.yaml
```

3. Run the worker (another terminal):
```bash
go run ./example/github/worker/main.go -config app.docker.yaml
```

4. Send a test webhook:
```bash
./scripts/send_webhook.sh github pull_request example/github/pull_request.json
```

Examples:
- `example/realworld` (multiple workers, single driver)
- `example/riverqueue` (publish to RiverQueue and consume with River workers)
- `example/gitlab` (GitLab webhook sample)
- `example/bitbucket` (Bitbucket webhook sample)

## Production Setup

For GitHub, prefer a GitHub App:
1. Create a GitHub App.
2. Set webhook URL to `https://<your-domain>/webhooks/github`.
3. Set `GITHUB_WEBHOOK_SECRET`.
4. Subscribe to required events only.
5. Deploy behind HTTPS.

## Configuration

Githooks is configured using a YAML file (local dev uses `app.docker.yaml`).

Further docs:
- `docs/drivers.md` (driver configuration)
- `docs/events.md` (provider event compatibility)
- `docs/rules.md` (rules engine)

### Providers
```yaml
providers:
  github:
    enabled: true
    path: /webhooks/github
    secret: ${GITHUB_WEBHOOK_SECRET}
  gitlab:
    enabled: false
    path: /webhooks/gitlab
    secret: ${GITLAB_WEBHOOK_SECRET} # optional
  bitbucket:
    enabled: false
    path: /webhooks/bitbucket
    secret: ${BITBUCKET_WEBHOOK_SECRET} # optional (X-Hook-UUID)
```

### Watermill Drivers (Publishers)

Driver fields (publisher config):
- `watermill.driver`: single driver.
- `watermill.drivers`: list of drivers to publish to (fan-out).
- `watermill.http`: publish-only (no subscriber).
- `watermill.sql`: requires DB driver import.

Driver fields (worker config):
- `watermill.driver`: single subscriber driver.
- `watermill.drivers`: subscribe to multiple drivers (fan-in). Unsupported drivers like `http` are skipped.
- `watermill.nats.client_id_suffix`: optional suffix for workers to avoid NATS Streaming client ID conflicts.

Example (AMQP):
```yaml
watermill:
  driver: amqp
  amqp:
    url: amqp://guest:guest@localhost:5672/
    mode: durable_queue
```

RiverQueue (Postgres):
```yaml
watermill:
  driver: riverqueue
  riverqueue:
    driver: postgres
    dsn: postgres://user:pass@localhost:5432/dbname?sslmode=disable
    table: river_job
    queue: default
    kind: githooks.event
    max_attempts: 25
    priority: 2
    tags: ["githooks", "webhook"]
```

Multiple drivers (fan-out):
```yaml
watermill:
  drivers: [amqp, http]
  amqp:
    url: amqp://guest:guest@localhost:5672/
    mode: durable_queue
  http:
    mode: base_url
    base_url: http://localhost:9000/hooks
```

### Rules

Rules use JSONPath + boolean logic:
```yaml
rules:
  - when: action == "opened" && pull_request.draft == false
    emit: pr.opened.ready
  - when: action == "closed" && pull_request.merged == true
    emit: pr.merged
    drivers: [amqp, http]
```

- Bare identifiers are treated as root JSONPath (e.g., `action` becomes `$.action`).
- `rules_strict: true` skips evaluation if any JSONPath in a rule is missing.
- If `drivers` is omitted, the event is published to all configured drivers.

## SDK (Workers)

Minimal example:
```go
sub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NewStdLogger(false, false))

wk := worker.New(
  worker.WithSubscriber(sub),
  worker.WithTopics("pr.opened.ready"),
  worker.WithConcurrency(10),
)

wk.HandleTopic("pr.opened.ready", func(ctx context.Context, evt *worker.Event) error {
  return nil
})

if err := wk.Run(ctx); err != nil {
  log.Fatal(err)
}
```

Watermill middleware adapter:
```go
import wm "github.com/ThreeDotsLabs/watermill/message/router/middleware"

wk := worker.New(
  worker.WithSubscriber(sub),
  worker.WithTopics("pr.opened.ready"),
  worker.WithMiddleware(worker.MiddlewareFromWatermill(wm.Retry{MaxRetries: 3}.Middleware)),
)
```

## Helm Charts

Helm charts live in `charts/`:
- `charts/githooks` deploys the webhook server.
- `charts/githooks-worker` deploys a worker.

Install from GitHub Pages:
```sh
helm repo add githooks https://yindia.github.io/githooks
helm repo update
helm install githooks githooks/githooks
helm install githooks-worker githooks/githooks-worker
```

## Releases

Code release:
- Tag `vX.Y.Z` to publish `ghcr.io/yindia/githooks` and Go module version.

Chart release:
- Update `charts/*/Chart.yaml` (`version` + `appVersion`).
- Tag `chart-X.Y.Z` to publish charts to `gh-pages`.

If a chart tag already exists, create a new tag or delete the GitHub release and re-run the workflow.

## Development

Run tests:
```bash
go test ./...
```

Notes:
- SQL publishing requires a DB driver import (`github.com/lib/pq` or `github.com/go-sql-driver/mysql`).
- Rules are evaluated in order; multiple matches publish multiple topics.
- Default test secret: `devsecret`.
