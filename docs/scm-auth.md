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

switch authCtx.Provider {
case "github":
	gh := client.(*github.Client)
	_, err = gh.GetRepo(ctx, "acme", "demo")
case "gitlab":
	gl := client.(*gitlab.Client)
	_, err = gl.GetRepo(ctx, "acme", "demo")
case "bitbucket":
	bb := client.(*bitbucket.Client)
	_, err = bb.GetRepo(ctx, "acme", "demo")
}
return err
```

## Notes

- GitHub uses GitHub App authentication. Tokens are short-lived and never persisted.
- GitLab and Bitbucket use access tokens from config.
