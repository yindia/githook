package auth

import (
	"context"
	"testing"
)

func TestResolverGitHubInstallation(t *testing.T) {
	cfg := Config{
		GitHub: ProviderConfig{
			AppID:          123,
			PrivateKeyPath: "/tmp/key.pem",
		},
	}
	resolver := NewResolver(cfg)

	payload := []byte(`{"installation":{"id":42}}`)
	authCtx, err := resolver.Resolve(context.Background(), EventContext{
		Provider: "github",
		Payload:  payload,
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if authCtx.InstallationID != 42 {
		t.Fatalf("expected installation id 42, got %d", authCtx.InstallationID)
	}
}

func TestResolverGitHubMissingInstallation(t *testing.T) {
	cfg := Config{
		GitHub: ProviderConfig{
			AppID:          123,
			PrivateKeyPath: "/tmp/key.pem",
		},
	}
	resolver := NewResolver(cfg)

	_, err := resolver.Resolve(context.Background(), EventContext{
		Provider: "github",
		Payload:  []byte(`{"installation":{}}`),
	})
	if err == nil {
		t.Fatalf("expected error for missing installation id")
	}
}

func TestResolverGitLabToken(t *testing.T) {
	cfg := Config{
		GitLab: ProviderConfig{
			Token: "glpat-123",
		},
	}
	resolver := NewResolver(cfg)

	authCtx, err := resolver.Resolve(context.Background(), EventContext{
		Provider: "gitlab",
		Payload:  []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if authCtx.Token != "glpat-123" {
		t.Fatalf("expected gitlab token")
	}
}

func TestResolverBitbucketToken(t *testing.T) {
	cfg := Config{
		Bitbucket: ProviderConfig{
			Token: "bb-123",
		},
	}
	resolver := NewResolver(cfg)

	authCtx, err := resolver.Resolve(context.Background(), EventContext{
		Provider: "bitbucket",
		Payload:  []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if authCtx.Token != "bb-123" {
		t.Fatalf("expected bitbucket token")
	}
}
