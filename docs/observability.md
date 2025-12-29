# Observability

Githooks exposes lightweight observability signals that work with minimal setup.

## Metrics (expvar)

Enable the built-in expvar endpoint to scrape basic counters.

```yaml
server:
  metrics_enabled: true
  metrics_path: /metrics
```

Metrics include:

- `githooks_requests_total`
- `githooks_parse_errors_total`
- `githooks_publish_errors_total`

## Request IDs

Incoming requests use or generate `X-Request-Id`. The server echoes it back in
responses and includes it in logs and published message metadata.

## Rate Limiting

The server supports a per-IP token-bucket limiter.

```yaml
server:
  rate_limit_rps: 10
  rate_limit_burst: 20
```

Set `rate_limit_rps` to `0` to disable rate limiting.
