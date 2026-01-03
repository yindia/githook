package worker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"githooks/pkg/providers/bitbucket"
	"githooks/pkg/providers/gitlab"
)

const defaultInstallationsBaseURL = "http://localhost:8080"

// ResolveProviderClient returns an authenticated provider client for the event.
// It uses the API base URL from GITHOOKS_API_BASE_URL or defaults to localhost.
func ResolveProviderClient(ctx context.Context, evt *Event) (interface{}, error) {
	baseURL := os.Getenv("GITHOOKS_API_BASE_URL")
	if baseURL == "" {
		configPath := os.Getenv("GITHOOKS_CONFIG_PATH")
		if configPath == "" {
			configPath = os.Getenv("GITHOOKS_CONFIG")
		}
		if configPath != "" {
			if cfg, err := LoadServerConfig(configPath); err == nil {
				baseURL = serverBaseURL(cfg)
			}
		}
	}
	if baseURL == "" {
		baseURL = defaultInstallationsBaseURL
	}
	return ResolveProviderClientWithClient(ctx, evt, &InstallationsClient{BaseURL: baseURL})
}

// ResolveProviderClientWithClient returns an authenticated provider client for the event.
// GitHub uses the injected client on the event. GitLab/Bitbucket fetch tokens via the InstallationsClient.
func ResolveProviderClientWithClient(ctx context.Context, evt *Event, client *InstallationsClient) (interface{}, error) {
	if evt == nil {
		return nil, errors.New("event is required")
	}

	switch evt.Provider {
	case "github":
		gh, ok := GitHubClient(evt)
		if !ok {
			return nil, errors.New("github client not available on event")
		}
		return gh, nil
	case "gitlab":
		record, err := ResolveInstallation(ctx, evt, client)
		if err != nil {
			return nil, err
		}
		if record == nil || record.AccessToken == "" {
			return nil, errors.New("gitlab access token missing")
		}
		return gitlab.NewClient(record.AccessToken)
	case "bitbucket":
		record, err := ResolveInstallation(ctx, evt, client)
		if err != nil {
			return nil, err
		}
		if record == nil || record.AccessToken == "" {
			return nil, errors.New("bitbucket access token missing")
		}
		return bitbucket.NewOAuthbearerToken(record.AccessToken), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", evt.Provider)
	}
}

func serverBaseURL(cfg ServerConfig) string {
	if strings.TrimSpace(cfg.PublicBaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	}
	port := cfg.Port
	if port == 0 {
		return ""
	}
	return "http://localhost:" + strconv.Itoa(port)
}
