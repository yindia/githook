package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"githooks/internal"
	"githooks/webhook"
)

func main() {
	appConfig, err := internal.LoadAppConfig("app.yaml")
	if err != nil {
		log.Fatalf("load app config: %v", err)
	}

	rulesConfig, err := internal.LoadRulesConfig("config.yaml")
	if err != nil {
		log.Fatalf("load rules config: %v", err)
	}

	ruleEngine, err := internal.NewRuleEngine(rulesConfig)
	if err != nil {
		log.Fatalf("compile rules: %v", err)
	}

	publisher, err := internal.NewPublisher(appConfig.Watermill)
	if err != nil {
		log.Fatalf("publisher: %v", err)
	}
	defer publisher.Close()

	mux := http.NewServeMux()

	if appConfig.Providers.GitHub.Enabled {
		ghHandler, err := webhook.NewGitHubHandler(
			appConfig.Providers.GitHub.Secret,
			ruleEngine,
			publisher,
		)
		if err != nil {
			log.Fatalf("github handler: %v", err)
		}
		mux.Handle(appConfig.Providers.GitHub.Path, ghHandler)
		log.Printf("github webhook enabled on %s", appConfig.Providers.GitHub.Path)
	}

	addr := ":" + strconv.Itoa(appConfig.Server.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
