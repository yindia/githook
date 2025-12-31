package worker

import (
	"context"
	"errors"

	"githooks/pkg/auth"
	"githooks/pkg/scm"
)

// SCMClientProvider resolves SCM clients from webhook events.
type SCMClientProvider struct {
	resolver auth.Resolver
	factory  *scm.Factory
}

// NewSCMClientProvider creates a provider that resolves auth and builds SCM clients.
func NewSCMClientProvider(cfg auth.Config) *SCMClientProvider {
	return &SCMClientProvider{
		resolver: auth.NewResolver(cfg),
		factory:  scm.NewFactory(cfg),
	}
}

// NewSCMClientProviderWithResolver creates a provider with custom resolver/factory.
func NewSCMClientProviderWithResolver(resolver auth.Resolver, factory *scm.Factory) *SCMClientProvider {
	return &SCMClientProvider{resolver: resolver, factory: factory}
}

// Client resolves a provider-specific SCM client for the given event.
func (p *SCMClientProvider) Client(ctx context.Context, evt *Event) (interface{}, error) {
	if p == nil || p.resolver == nil || p.factory == nil {
		return nil, errors.New("scm client provider is not configured")
	}
	if evt == nil {
		return nil, errors.New("event is required")
	}
	authCtx, err := p.resolver.Resolve(ctx, auth.EventContext{
		Provider: evt.Provider,
		Payload:  evt.Payload,
	})
	if err != nil {
		return nil, err
	}
	return p.factory.NewClient(ctx, authCtx)
}
