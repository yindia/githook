# Driver Configuration

This document covers Watermill driver configuration for publishers and workers.

## Publisher Drivers

Use `watermill.driver` for a single driver or `watermill.drivers` for fan-out.

Example (single driver):
```yaml
watermill:
  driver: amqp
  amqp:
    url: amqp://guest:guest@localhost:5672/
    mode: durable_queue
```

Example (fan-out):
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

## Worker Drivers

Workers subscribe via `watermill.driver` (single) or `watermill.drivers` (multi). Unsupported drivers like `http` are skipped.

Example (single driver):
```yaml
watermill:
  driver: amqp
  amqp:
    url: amqp://guest:guest@localhost:5672/
    mode: durable_queue
```

Example (multi-driver fan-in):
```yaml
watermill:
  drivers: [amqp, nats, kafka, sql]
```

## Driver Reference

### GoChannel
```yaml
watermill:
  driver: gochannel
  gochannel:
    output_buffer: 64
    persistent: false
    block_publish_until_subscriber_ack: false
```

### Kafka
```yaml
watermill:
  driver: kafka
  kafka:
    brokers: ["localhost:9092"]
```

### NATS Streaming
```yaml
watermill:
  driver: nats
  nats:
    cluster_id: test-cluster
    client_id: githooks
    client_id_suffix: "-worker" # workers only
    url: nats://localhost:4222
```

### AMQP (RabbitMQ)
```yaml
watermill:
  driver: amqp
  amqp:
    url: amqp://guest:guest@localhost:5672/
    mode: durable_queue
```

### SQL (Postgres/MySQL)
```yaml
watermill:
  driver: sql
  sql:
    driver: postgres # or mysql
    dsn: postgres://user:pass@localhost:5432/dbname?sslmode=disable
    dialect: postgres # or mysql
    auto_initialize_schema: true
```

### RiverQueue (Postgres)
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

### HTTP (Publish-only)
```yaml
watermill:
  driver: http
  http:
    mode: base_url
    base_url: http://localhost:9000/hooks
```

Notes:
- HTTP is publish-only (no worker subscriber).
- SQL publishing requires importing a DB driver (e.g., `github.com/lib/pq`).
