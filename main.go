package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"githooks/internal"
	"githooks/pkg/auth"
	"githooks/pkg/api"
	"githooks/pkg/oauth"
	"githooks/pkg/storage/installations"
	"githooks/pkg/storage/namespaces"
	"githooks/pkg/webhook"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	logger := internal.NewLogger("server")
	configPath := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	config, err := internal.LoadConfig(*configPath)
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	ruleEngine, err := internal.NewRuleEngine(internal.RulesConfig{
		Rules:  config.Rules,
		Strict: config.RulesStrict,
		Logger: logger,
	})
	if err != nil {
		logger.Fatalf("compile rules: %v", err)
	}

	publisher, err := internal.NewPublisher(config.Watermill)
	if err != nil {
		logger.Fatalf("publisher: %v", err)
	}
	defer publisher.Close()

	var installStore *installations.Store
	var namespaceStore *namespaces.Store
	if config.Storage.Driver != "" && config.Storage.DSN != "" {
		store, err := installations.Open(installations.Config{
			Driver:      config.Storage.Driver,
			DSN:         config.Storage.DSN,
			Dialect:     config.Storage.Dialect,
			AutoMigrate: config.Storage.AutoMigrate,
		})
		if err != nil {
			logger.Fatalf("storage: %v", err)
		}
		installStore = store
		defer installStore.Close()
		logger.Printf("storage enabled driver=%s dialect=%s table=githooks_installations", config.Storage.Driver, config.Storage.Dialect)

		nsStore, err := namespaces.Open(namespaces.Config{
			Driver:      config.Storage.Driver,
			DSN:         config.Storage.DSN,
			Dialect:     config.Storage.Dialect,
			AutoMigrate: config.Storage.AutoMigrate,
		})
		if err != nil {
			logger.Fatalf("namespaces storage: %v", err)
		}
		namespaceStore = nsStore
		defer namespaceStore.Close()
		logger.Printf("namespaces enabled driver=%s dialect=%s table=git_namespaces", config.Storage.Driver, config.Storage.Dialect)
	} else {
		logger.Printf("storage disabled (missing storage.driver or storage.dsn)")
	}

	mux := http.NewServeMux()
	mux.Handle("/", &oauth.StartHandler{
		Providers:     config.Providers,
		PublicBaseURL: config.Server.PublicBaseURL,
		Logger:        logger,
	})
	mux.Handle("/api/installations", &api.InstallationsHandler{
		Store:  installStore,
		Logger: logger,
	})
	mux.Handle("/api/namespaces", &api.NamespacesHandler{
		Store:  namespaceStore,
		Logger: logger,
	})
	mux.Handle("/api/namespaces/sync", &api.SyncNamespacesHandler{
		InstallStore:  installStore,
		NamespaceStore: namespaceStore,
		Providers:     config.Providers,
		Logger:        logger,
	})
	mux.Handle("/api/webhooks/namespace", &api.NamespaceWebhookHandler{
		Store:         namespaceStore,
		InstallStore:  installStore,
		Providers:     config.Providers,
		PublicBaseURL: config.Server.PublicBaseURL,
		Logger:        logger,
	})

	if config.Providers.GitHub.Enabled {
		ghHandler, err := webhook.NewGitHubHandler(
			config.Providers.GitHub.Secret,
			ruleEngine,
			publisher,
			logger,
			config.Server.MaxBodyBytes,
			config.Server.DebugEvents,
			installStore,
			namespaceStore,
		)
		if err != nil {
			logger.Fatalf("github handler: %v", err)
		}
		mux.Handle(config.Providers.GitHub.Path, ghHandler)
		logger.Printf(
			"provider=github webhook=enabled path=%s oauth_callback=/oauth/github/callback app_id=%d",
			config.Providers.GitHub.Path,
			config.Providers.GitHub.AppID,
		)
	}

	if config.Providers.GitLab.Enabled {
		glHandler, err := webhook.NewGitLabHandler(
			config.Providers.GitLab.Secret,
			ruleEngine,
			publisher,
			logger,
			config.Server.MaxBodyBytes,
			config.Server.DebugEvents,
			namespaceStore,
		)
		if err != nil {
			logger.Fatalf("gitlab handler: %v", err)
		}
		mux.Handle(config.Providers.GitLab.Path, glHandler)
		logger.Printf(
			"provider=gitlab webhook=enabled path=%s oauth_callback=/oauth/gitlab/callback",
			config.Providers.GitLab.Path,
		)
	}

	if config.Providers.Bitbucket.Enabled {
		bbHandler, err := webhook.NewBitbucketHandler(
			config.Providers.Bitbucket.Secret,
			ruleEngine,
			publisher,
			logger,
			config.Server.MaxBodyBytes,
			config.Server.DebugEvents,
			namespaceStore,
		)
		if err != nil {
			logger.Fatalf("bitbucket handler: %v", err)
		}
		mux.Handle(config.Providers.Bitbucket.Path, bbHandler)
		logger.Printf(
			"provider=bitbucket webhook=enabled path=%s oauth_callback=/oauth/bitbucket/callback",
			config.Providers.Bitbucket.Path,
		)
	}

	redirectBase := config.OAuth.RedirectBaseURL
	oauthHandler := func(provider string, cfg auth.ProviderConfig) *oauth.Handler {
		return &oauth.Handler{
			Provider:     provider,
			Config:       cfg,
			Providers:    config.Providers,
			Store:        installStore,
			NamespaceStore: namespaceStore,
			Logger:       logger,
			RedirectBase: redirectBase,
			PublicBaseURL: config.Server.PublicBaseURL,
		}
	}
	mux.Handle("/oauth/github/callback", oauthHandler("github", config.Providers.GitHub))
	mux.Handle("/oauth/gitlab/callback", oauthHandler("gitlab", config.Providers.GitLab))
	mux.Handle("/oauth/bitbucket/callback", oauthHandler("bitbucket", config.Providers.Bitbucket))

	handler := h2c.NewHandler(mux, &http2.Server{})

	addr := ":" + strconv.Itoa(config.Server.Port)
	if config.Server.PublicBaseURL != "" {
		logger.Printf("server public_base_url=%s", config.Server.PublicBaseURL)
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       time.Duration(config.Server.ReadTimeoutMS) * time.Millisecond,
		WriteTimeout:      time.Duration(config.Server.WriteTimeoutMS) * time.Millisecond,
		IdleTimeout:       time.Duration(config.Server.IdleTimeoutMS) * time.Millisecond,
		ReadHeaderTimeout: time.Duration(config.Server.ReadHeaderMS) * time.Millisecond,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Printf("listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %v", err)
		}
	}()

	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Printf("shutdown: %v", err)
	}
}
