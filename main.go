package main

import (
	"context"
	"expvar"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"githooks/internal"
	"githooks/pkg/webhook"

	_ "github.com/lib/pq"
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

	mux := http.NewServeMux()

	if config.Providers.GitHub.Enabled {
		ghHandler, err := webhook.NewGitHubHandler(
			config.Providers.GitHub.Secret,
			ruleEngine,
			publisher,
			logger,
			config.Server.MaxBodyBytes,
		)
		if err != nil {
			logger.Fatalf("github handler: %v", err)
		}
		mux.Handle(config.Providers.GitHub.Path, ghHandler)
		logger.Printf("github webhook enabled on %s", config.Providers.GitHub.Path)
	}

	if config.Providers.GitLab.Enabled {
		glHandler, err := webhook.NewGitLabHandler(
			config.Providers.GitLab.Secret,
			ruleEngine,
			publisher,
			logger,
			config.Server.MaxBodyBytes,
		)
		if err != nil {
			logger.Fatalf("gitlab handler: %v", err)
		}
		mux.Handle(config.Providers.GitLab.Path, glHandler)
		logger.Printf("gitlab webhook enabled on %s", config.Providers.GitLab.Path)
	}

	if config.Providers.Bitbucket.Enabled {
		bbHandler, err := webhook.NewBitbucketHandler(
			config.Providers.Bitbucket.Secret,
			ruleEngine,
			publisher,
			logger,
			config.Server.MaxBodyBytes,
		)
		if err != nil {
			logger.Fatalf("bitbucket handler: %v", err)
		}
		mux.Handle(config.Providers.Bitbucket.Path, bbHandler)
		logger.Printf("bitbucket webhook enabled on %s", config.Providers.Bitbucket.Path)
	}

	if config.Server.MetricsEnabled {
		mux.Handle(config.Server.MetricsPath, expvar.Handler())
	}

	handler := http.Handler(mux)
	if config.Server.RateLimitRPS > 0 {
		handler = internal.NewRateLimitHandler(handler, config.Server.RateLimitRPS, config.Server.RateLimitBurst, 5*time.Minute)
	}

	addr := ":" + strconv.Itoa(config.Server.Port)
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
