package github

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client is the official GitHub SDK client.
type Client = gh.Client

// NewAppClient creates a GitHub SDK client by exchanging an installation token.
func NewAppClient(ctx context.Context, cfg AppConfig, installationID int64) (*Client, error) {
	if installationID == 0 {
		return nil, fmt.Errorf("github installation id is required")
	}
	authenticator := newAppAuthenticator(cfg)
	token, err := authenticator.installationToken(ctx, installationID)
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, ts)

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL != "" && baseURL != defaultBaseURL {
		client, err := gh.NewEnterpriseClient(baseURL, baseURL, httpClient)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	return gh.NewClient(httpClient), nil
}
