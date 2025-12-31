package auth

import (
	"context"
	"errors"
	"strings"

	"githooks/pkg/providers/github"
)

// AuthContext contains the resolved authentication data for a webhook event.
type AuthContext struct {
	Provider       string
	InstallationID int64
	Token          string
}

// EventContext captures the webhook context used for auth resolution.
type EventContext struct {
	Provider string
	Payload  []byte
}

// Resolver resolves authentication for a webhook event.
type Resolver interface {
	Resolve(ctx context.Context, event EventContext) (AuthContext, error)
}

// DefaultResolver resolves auth using configuration and webhook payload data.
type DefaultResolver struct {
	cfg Config
}

// NewResolver constructs a DefaultResolver.
func NewResolver(cfg Config) *DefaultResolver {
	return &DefaultResolver{cfg: cfg}
}

// Resolve builds an AuthContext from the webhook event.
func (r *DefaultResolver) Resolve(_ context.Context, event EventContext) (AuthContext, error) {
	provider := strings.ToLower(strings.TrimSpace(event.Provider))
	switch provider {
	case "github":
		if r.cfg.GitHub.AppID == 0 || r.cfg.GitHub.PrivateKeyPath == "" {
			return AuthContext{}, errors.New("github app_id and private_key_path are required")
		}
		installationID, ok, err := github.InstallationIDFromPayload(event.Payload)
		if err != nil {
			return AuthContext{}, err
		}
		if !ok {
			return AuthContext{}, errors.New("github installation id not found in payload")
		}
		return AuthContext{
			Provider:       "github",
			InstallationID: installationID,
		}, nil
	case "gitlab":
		if r.cfg.GitLab.Token == "" {
			return AuthContext{}, errors.New("gitlab token is required")
		}
		return AuthContext{
			Provider: "gitlab",
			Token:    r.cfg.GitLab.Token,
		}, nil
	case "bitbucket":
		if r.cfg.Bitbucket.Token == "" {
			return AuthContext{}, errors.New("bitbucket token is required")
		}
		return AuthContext{
			Provider: "bitbucket",
			Token:    r.cfg.Bitbucket.Token,
		}, nil
	default:
		return AuthContext{}, errors.New("unsupported provider for auth resolution")
	}
}
