Githooks

Config-driven webhook router for GitHub (with GitLab/Bitbucket planned). It normalizes inbound webhook events, evaluates them against YAML rules, and publishes matching events to Watermill topics for downstream consumers.

Features
- Typed webhook parsing via go-playground/webhooks
- Provider-agnostic normalized event model
- Rule-based routing via govaluate
- Watermill-backed publishing (gochannel, Kafka, NATS Streaming, AMQP, SQL)
- Stateless by default

Architecture
Webhook Provider -> go-playground/webhooks -> Adapter -> Normalized Event -> Rule Engine -> Watermill Publisher

Quickstart
1) Configure secrets and rules in `app.yaml` and `config.yaml`.
2) Export any secrets referenced by env vars.
3) Run:
   go run main.go

Example
export GITHUB_WEBHOOK_SECRET=devsecret
go run main.go

Then send GitHub webhooks to:
http://localhost:8080/webhooks/github

Configuration
`app.yaml`
server:
  port: 8080

providers:
  github:
    enabled: true
    path: /webhooks/github
    secret: ${GITHUB_WEBHOOK_SECRET}

watermill:
  driver: gochannel

`config.yaml`
rules:
  - when: action == "opened" && draft == false
    emit: pr.opened.ready
  - when: action == "closed" && merged == true
    emit: pr.merged

Normalized event model
Provider: github, gitlab, bitbucket
Name:     pull_request, push, ...
Data:     flattened payload fields used by rules

Watermill drivers
gochannel:
  watermill:
    driver: gochannel
    gochannel:
      output_buffer: 64
      persistent: false
      block_publish_until_subscriber_ack: false

kafka:
  watermill:
    driver: kafka
    kafka:
      brokers: ["localhost:9092"]

nats (streaming):
  watermill:
    driver: nats
    nats:
      cluster_id: test-cluster
      client_id: githooks
      url: nats://localhost:4222

amqp:
  watermill:
    driver: amqp
    amqp:
      url: amqp://guest:guest@localhost:5672/
      mode: durable_queue

sql:
  watermill:
    driver: sql
    sql:
      driver: postgres
      dsn: postgres://user:pass@localhost:5432/dbname?sslmode=disable
      dialect: postgres
      auto_initialize_schema: true

Notes
- SQL publishing requires a database driver import (e.g., lib/pq or go-sql-driver/mysql) in your app.
- Rules are evaluated in order; multiple matches publish multiple topics.

Testing
go test ./...
