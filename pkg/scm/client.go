package scm

import (
	"context"
	"errors"

	"githooks/pkg/auth"
	"githooks/pkg/providers/bitbucket"
	"githooks/pkg/providers/github"
	"githooks/pkg/providers/gitlab"
)

// SCMClient is a provider-specific API client.
// Use type assertions to access the provider client you need.
type SCMClient interface{}

// Factory builds SCM clients using resolved auth contexts.
type Factory struct {
	cfg auth.Config
}

// NewFactory creates a new Factory.
func NewFactory(cfg auth.Config) *Factory {
	return &Factory{cfg: cfg}
}

// NewClient creates a provider-specific client from an AuthContext.
func (f *Factory) NewClient(ctx context.Context, authCtx auth.AuthContext) (SCMClient, error) {
	switch authCtx.Provider {
	case "github":
		return github.NewAppClient(ctx, github.AppConfig{
			AppID:          f.cfg.GitHub.AppID,
			PrivateKeyPath: f.cfg.GitHub.PrivateKeyPath,
			BaseURL:        f.cfg.GitHub.BaseURL,
		}, authCtx.InstallationID)
	case "gitlab":
		return gitlab.NewTokenClient(f.cfg.GitLab, authCtx.Token), nil
	case "bitbucket":
		return bitbucket.NewTokenClient(f.cfg.Bitbucket, authCtx.Token), nil
	default:
		return nil, errors.New("unsupported provider for scm client")
	}
}
