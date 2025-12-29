# Event Compatibility

Githooks preserves provider event names in `Event.Name` and sets `Event.Provider` to the source system.

## GitHub
- Header: `X-GitHub-Event`
- Signature: `X-Hub-Signature` (HMAC SHA-1)
- Path: `/webhooks/github`

## GitLab
- Header: `X-Gitlab-Event`
- Secret: `X-Gitlab-Token` (optional)
- Path: `/webhooks/gitlab`

## Bitbucket (Cloud)
- Header: `X-Event-Key`
- Secret: `X-Hook-UUID` (optional)
- Path: `/webhooks/bitbucket`

## Compatibility Notes
- GitHub payloads use `pull_request` (singular), not `pull_requests`.
- Bitbucket events use keys like `pullrequest:created`.
- GitLab event names come from `X-Gitlab-Event` (e.g., `Merge Request Hook`).

## Debugging
Check logs for:
- `event provider=... name=... topics=[...]`
- `rule debug: when=... params=...`
