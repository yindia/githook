# Githooks

Githooks is a config-driven webhook router and worker SDK for GitHub, GitLab, and Bitbucket. It normalizes inbound webhook events, evaluates them against YAML rules, and publishes matching events to [Watermill](https://watermill.io/) topics. The SDK then lets you build broker-agnostic workers with client injection, retries, and graceful shutdown.

**Warning:** This project is intended for research and development use only. It is not production-ready.

## Features


- **Multi-Provider Support**: Handles webhooks from GitHub, GitLab, and Bitbucket.
- **Normalized Events**: Converts provider-specific payloads into a consistent, easy-to-use format.
- **Powerful Rule Engine**: Filter and route events using YAML-based rules with JSONPath and boolean logic.
- **Flexible Publishing**: Publish events to multiple message queues and protocols via Watermill (GoChannel, Kafka, NATS, AMQP, SQL, HTTP, and more).
- **Stateless & Scalable**: The server is stateless and can be scaled horizontally.
- **Simple Worker SDK**: A lightweight SDK for building event consumers.

## Table of Contents

- [Architecture](#architecture)
- [Getting Started (Local)](#getting-started-local)
- [Configuration](#configuration)
  - [Providers](#providers)
  - [Watermill Drivers (Publishing)](#watermill-drivers-publishing)
  - [Rules](#rules)
- [Worker SDK](#worker-sdk)
- [Examples](#examples)
- [Helm Charts](#helm-charts)
- [Releases](#releases)
- [Development](#development)

## Architecture

The server ingests webhooks, normalizes them, and evaluates them against user-defined rules. Matching events are then published to one or more Watermill topics.

```mermaid
flowchart LR
  A[Webhook Provider\nGitHub/GitLab/Bitbucket] --> B[Webhook Server]
  B -- Parse --> C(Provider Adapter)
  C -- Normalize --> D(Event)
  D -- Evaluate --> E{Rule Engine}
  E -- Match --> F[Watermill Publisher]
  F -- Publish --> G[AMQP/Kafka/NATS/SQL/HTTP]
  G --> H[Worker/Consumer]
```

## Getting Started (Local)

1.  **Start dependencies:**

    ```bash
    docker compose up -d
    ```

2.  **Run the server:**

    Set the secret for validating GitHub webhooks and run the server with the local Docker config.

    ```bash
    export GITHUB_WEBHOOK_SECRET=devsecret
    go run ./main.go -config app.docker.yaml
    ```

3.  **Run a worker:**

    In another terminal, run an example worker that listens for events.

    ```bash
    go run ./example/github/worker/main.go -config app.docker.yaml
    ```

4.  **Send a test webhook:**

    Use the provided script to simulate a GitHub `pull_request` event.

    ```bash
    ./scripts/send_webhook.sh github pull_request example/github/pull_request.json
    ```

    You should see the server log the event and the worker log its "pr.opened.ready" message.

## Configuration

Docs:
- [Driver configuration](docs/drivers.md)
- [Event compatibility](docs/events.md)
- [Getting started (GitHub)](docs/getting-started-github.md)
- [Getting started (GitLab)](docs/getting-started-gitlab.md)
- [Getting started (Bitbucket)](docs/getting-started-bitbucket.md)
- [Rules engine](docs/rules.md)
- [Observability](docs/observability.md)
- [Webhook setup](docs/webhooks.md)
- [SDK client injection](docs/sdk_clients.md)

Githooks is configured using a single YAML file. Environment variables like `${VAR}` are automatically expanded.
Requests use or generate `X-Request-Id`, which is echoed back in responses and included in logs.

### Providers

The `providers` section configures the webhook endpoints for each Git provider.

```yaml
providers:
  github:
    enabled: true
    path: /webhooks/github
    secret: ${GITHUB_WEBHOOK_SECRET}
  gitlab:
    enabled: false
    path: /webhooks/gitlab
    secret: ${GITLAB_WEBHOOK_SECRET} # Optional
  bitbucket:
    enabled: false
    path: /webhooks/bitbucket
    secret: ${BITBUCKET_WEBHOOK_SECRET} # Optional, for X-Hook-UUID
```

### Server Limits

```yaml
server:
  port: 8080
  read_timeout_ms: 5000
  write_timeout_ms: 10000
  idle_timeout_ms: 60000
  read_header_timeout_ms: 5000
  max_body_bytes: 1048576
  rate_limit_rps: 0
  rate_limit_burst: 0
  metrics_enabled: false
  metrics_path: /metrics
```

### Watermill Drivers (Publishing)

The `watermill` section configures the message broker(s) to publish events to.

-   `driver`: (string) Default publisher driver.
-   `drivers`: (array) Fan-out to all listed drivers by default.

**Single Driver (AMQP)**
```yaml
watermill:
  driver: amqp
  amqp:
    url: amqp://guest:guest@localhost:5672/
    mode: durable_queue # Or: nondurable_queue, durable_pubsub, nondurable_pubsub
```

**Multiple Drivers (Fan-Out)**
```yaml
watermill:
  drivers: [amqp, http]
  amqp:
    url: amqp://guest:guest@localhost:5672/
  http:
    mode: base_url
    base_url: http://localhost:9000/hooks
```

**RiverQueue (Postgres Job Queue)**
```yaml
watermill:
  driver: riverqueue
  riverqueue:
    driver: postgres
    dsn: postgres://user:pass@localhost:5432/dbname?sslmode=disable
    table: river_job # Optional, default is river_job
    queue: default   # Optional, default is default
    kind: githooks.event # The job type to insert
```

See the [Watermill documentation](https://watermill.io/docs/pub-subs/) for details on each driver's configuration.

### Rules

The `rules` section defines which events to publish and where. Each rule has a `when` condition and an `emit` topic.

```yaml
rules_strict: false # Optional: if true, rules with missing fields won't match
rules:
  # If a PR is opened and not a draft, emit to 'pr.opened.ready'
  - when: action == "opened" && pull_request.draft == false
    emit: pr.opened.ready

  # If a PR is merged, emit to 'pr.merged' on specific drivers
  - when: action == "closed" && pull_request.merged == true
    emit: pr.merged
    drivers: [amqp, http]
```

-   **`when`**: A boolean expression evaluated against the webhook payload.
    -   Bare identifiers (e.g., `action`) are treated as JSONPath `$.action`.
    -   You can use full JSONPath syntax (e.g., `$.pull_request.head.ref`).
-   **`emit`**: The topic name to publish the event to if the `when` condition is true.
-   **`drivers`**: (Optional) A list of specific drivers to publish this event to. If omitted, the default `driver` or `drivers` from the Watermill config are used.

## Worker SDK

The worker SDK provides a simple way to consume events from the message broker.

**Minimal Example**
```go
package main

import (
    "context"
    "log"

    "github.com/ThreeDotsLabs/watermill"
    "github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
    "githooks/pkg/worker"
)

func main() {
    // In a real app, you would configure a persistent subscriber (e.g., Kafka, AMQP)
    sub, err := worker.BuildSubscriber(cfg.Watermill)
    if err != nil {
        log.Fatalf("Failed to build subscriber: %v", err)
    }

    wk := worker.New(
        worker.WithSubscriber(sub),
        worker.WithTopics("pr.opened.ready"), // List of topics to subscribe to
        worker.WithConcurrency(10),
    )

    // Register a handler for a specific topic
    wk.HandleTopic("pr.opened.ready", func(ctx context.Context, evt *worker.Event) error {
        log.Printf("Received event: %s/%s", evt.Provider, evt.Type)
        // Do something with evt.Payload or evt.Normalized
        return nil
    })

    // Run the worker (blocking call)
    if err := wk.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

**Watermill Middleware**

You can use any Watermill middleware with the provided adapter.

```go
import wmmw "github.com/ThreeDotsLabs/watermill/message/router/middleware"

retryMiddleware := worker.MiddlewareFromWatermill(
    wmmw.Retry{MaxRetries: 3}.Middleware,
)

wk := worker.New(
  // ... other options
  worker.WithMiddleware(retryMiddleware),
)
```

## Building Your Own Go App

If you like this model of Git provider webhook management, you can build your own Go app by reusing the same pattern: validate provider signatures, normalize payloads, evaluate rules, then publish to a broker and consume with workers. Use the SDK to wire provider clients into handlers and keep business logic isolated from transport.

## Examples

The `example/` directory contains several working examples:
-   **`example/github`**: A simple server and worker for handling GitHub webhooks.
-   **`example/realworld`**: A more complex setup with multiple workers consuming events from a single topic.
-   **`example/riverqueue`**: Demonstrates publishing events to a [River](https://github.com/riverqueue/river) job queue.
-   **`example/gitlab`**: Sample setup for GitLab webhooks.
-   **`example/bitbucket`**: Sample setup for Bitbucket webhooks.

## Helm Charts

Helm charts for deploying the server and a generic worker are available in the `charts/` directory.

**Install from GitHub Pages**
```sh
helm repo add githooks https://yindia.github.io/githooks
helm repo update
helm install my-githooks githooks/githooks
helm install my-worker githooks/githooks-worker
```

## Releases

-   **Code Releases**: Tagging a commit with `vX.Y.Z` triggers a workflow that publishes a new Go module version and a container image to `ghcr.io/yindia/githooks`.
-   **Chart Releases**: Tagging a commit with `chart-X.Y.Z` publishes the Helm charts to the `gh-pages` branch. Ensure you update `version` and `appVersion` in `charts/*/Chart.yaml` first.

## Development

**Run Tests**
```bash
go test ./...
```

**Notes**
-   When using the SQL publisher, you must blank-import a database driver (e.g., `_ "github.com/lib/pq"`).
-   The default webhook secret for local testing is `devsecret`.
-   Rules are evaluated in the order they appear in the config file. Multiple rules can match a single event, causing multiple messages to be published.
