# SCM Authentication

This guide covers runtime SCM authentication resolved from webhook context.

## Configuration

```yaml
providers:
  github:
    app_id: 123
    private_key_path: /secrets/github.pem
    base_url: https://api.github.com
  gitlab:
    token: glpat-xxxx
    base_url: https://gitlab.com/api/v4
  bitbucket:
    token: bb-xxxx
    base_url: https://api.bitbucket.org/2.0
```

## Flow

1. Webhook arrives and is parsed (unchanged).
2. Resolve auth from the webhook payload.
3. Create an SCM client and call the provider API.

## Example: webhook → auth → API call

```go
resolver := auth.NewResolver(cfg.Providers)
authCtx, err := resolver.Resolve(ctx, auth.EventContext{
	Provider: evt.Provider,
	Payload:  evt.Payload,
})
if err != nil {
	return err
}

factory := scm.NewFactory(cfg.Providers)
client, err := factory.NewClient(ctx, authCtx)
if err != nil {
	return err
}

_, err = client.GetRepo(ctx, "acme", "demo")
return err
```

## Notes

- GitHub uses GitHub App authentication. Tokens are short-lived and never persisted.
- GitLab and Bitbucket use access tokens from config.
