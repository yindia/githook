package controllers

import (
	"context"
	"log"

	"githooks/pkg/worker"
)

func HandlePullRequestReady(ctx context.Context, evt *worker.Event) error {
	if gh, ok := worker.GitHubClient(evt); ok {
		_ = gh
	}
	log.Printf("topic=%s provider=%s", evt.Topic, evt.Provider)
	return nil
}

func HandlePullRequestMerged(ctx context.Context, evt *worker.Event) error {
	if gh, ok := worker.GitHubClient(evt); ok {
		_ = gh
	}
	log.Printf("topic=%s provider=%s", evt.Topic, evt.Provider)
	return nil
}
