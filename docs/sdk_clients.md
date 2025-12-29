# SDK Client Injection

Use the SDK to attach provider-specific clients to each event. Build the app clients once, then inject them per event with `ProviderClients` (or `ClientProviderFunc`).

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
