# SDK Client Injection

Use the SDK to attach provider-specific clients to each event. You can either inject your own clients or let the SDK resolve them from the webhook payload using the `providers` config.

## Pattern

```go
githubClient := newGitHubAppClient(appID, installationID, privateKeyPEM)
gitlabClient := newGitLabClient(token)
bitbucketClient := newBitbucketClient(username, appPassword)

wk := worker.New(
  worker.WithSubscriber(sub),
  worker.WithTopics("pr.opened.ready", "pr.merged"),
  worker.WithClientProvider(worker.ProviderClients{
    GitHub: func(ctx context.Context, evt *worker.Event) (interface{}, error) { return githubClient, nil },
    GitLab: func(ctx context.Context, evt *worker.Event) (interface{}, error) { return gitlabClient, nil },
    Bitbucket: func(ctx context.Context, evt *worker.Event) (interface{}, error) { return bitbucketClient, nil },
  }),
)

wk.HandleTopic("pr.opened.ready", func(ctx context.Context, evt *worker.Event) error {
  switch evt.Provider {
  case "github":
    gh := evt.Client.(*github.Client)
    _ = gh
  case "gitlab":
    gl := evt.Client.(*gitlab.Client)
    _ = gl
  case "bitbucket":
    bb := evt.Client.(*bitbucket.Client)
    _ = bb
  }
  return nil
})
```

This keeps webhook payloads normalized in `evt.Normalized`, while the SDK gives you the correct provider client for API calls.

## Auto-resolve clients from config

```go
wk := worker.New(
  worker.WithSubscriber(sub),
  worker.WithTopics("pr.opened.ready", "pr.merged"),
  worker.WithClientProvider(worker.NewSCMClientProvider(cfg.Providers)),
)
```

The `providers` section in your config includes the SCM auth settings (`app_id`, `private_key_path`, tokens, and base URLs).
