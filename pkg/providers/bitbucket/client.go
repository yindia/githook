package bitbucket

import (
	"errors"
	"os"
	"strings"

	"githooks/pkg/auth"

	bb "github.com/ktrysmt/go-bitbucket"
)

// Client is the Bitbucket SDK client.
type Client = bb.Client

// NewTokenClient returns a Bitbucket SDK client using an OAuth bearer token.
func NewTokenClient(cfg auth.ProviderConfig, token string) (*Client, error) {
	if token == "" {
		token = cfg.Token
	}
	if token == "" {
		return nil, errors.New("bitbucket token is required")
	}
	if base := normalizeBaseURL(cfg.BaseURL); base != "" {
		_ = os.Setenv("BITBUCKET_API_BASE_URL", base)
	}
	return bb.NewOAuthbearerToken(token)
}

func normalizeBaseURL(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return ""
	}
	return base
}
