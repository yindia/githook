# Event Compatibility

Githooks normalizes provider payloads but preserves provider event names in `Event.Name` and `Event.Provider`.

## GitHub
- Header: `X-GitHub-Event`
- Secret header: `X-Hub-Signature` (SHA-1)
- Path: `/webhooks/github`

## GitLab
- Header: `X-Gitlab-Event`
- Secret header: `X-Gitlab-Token` (optional)
- Path: `/webhooks/gitlab`

## Bitbucket (Cloud)
- Header: `X-Event-Key`
- Secret header: `X-Hook-UUID` (optional)
- Path: `/webhooks/bitbucket`

## Compatibility

Rules are evaluated against normalized payloads. The incoming payload must match the provider’s expected structure.

Common pitfalls:
- GitHub uses `pull_request` (singular), not `pull_requests`.
- Bitbucket event names look like `pullrequest:created`.
- GitLab event names look like `Merge Request Hook` (from the header).

Use `event provider=... name=... topics=[...]` logs to confirm routing.
