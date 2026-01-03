# Rules Engine

Rules use JSONPath for field access and boolean logic for matching. Each matching rule emits a topic and can optionally target specific drivers.

## Syntax
```yaml
rules:
  - when: action == "opened" && pull_request.draft == false
    emit: pr.opened.ready
  - when: action == "closed" && pull_request.merged == true
    emit: pr.merged
    drivers: [amqp, http]
```

## JSONPath
- Bare identifiers are treated as root JSONPath (e.g., `action` becomes `$.action`).
- Arrays are supported: `$.pull_request.commits[0].created == true`.

## Driver Targeting
- `drivers` omitted: publish to all configured drivers.
- `drivers` specified: publish only to those drivers.

## Strict Mode
Set `rules_strict: true` to skip a rule if any JSONPath in its `when` clause is missing.

## System Rules (GitHub App)
GitHub App installation events are always processed to keep `githooks_installations` in sync.
These updates are applied even if no user rule matches and cannot be disabled by rules:

- `installation`
- `installation_repositories`
