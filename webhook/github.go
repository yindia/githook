package webhook

import (
	"encoding/json"
	"log"
	"net/http"

	"githooks/internal"

	"github.com/go-playground/webhooks/v6/github"
)

type GitHubHandler struct {
	hook      *github.Webhook
	rules     *internal.RuleEngine
	publisher internal.Publisher
}

func NewGitHubHandler(secret string, rules *internal.RuleEngine, publisher internal.Publisher) (*GitHubHandler, error) {
	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return nil, err
	}

	return &GitHubHandler{hook: hook, rules: rules, publisher: publisher}, nil
}

func (h *GitHubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := h.hook.Parse(
		r,
		github.PingEvent,
		github.PullRequestEvent,
		github.PushEvent,
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch event := payload.(type) {
	case github.PingPayload:
		w.WriteHeader(http.StatusOK)
		return
	case github.PullRequestPayload:
		data := map[string]interface{}{
			"action": event.Action,
			"draft":  event.PullRequest.Draft,
			"merged": event.PullRequest.Merged,
			"base":   event.PullRequest.Base.Ref,
			"ref":    event.PullRequest.Head.Ref,
		}
		h.emit(r, internal.Event{Provider: "github", Name: "pull_request", Data: data})
	case github.PushPayload:
		data := map[string]interface{}{
			"ref":     event.Ref,
			"created": event.Created,
			"deleted": event.Deleted,
			"forced":  event.Forced,
			"size":    len(event.Commits),
		}
		h.emit(r, internal.Event{Provider: "github", Name: "push", Data: data})
	default:
		data, marshalErr := jsonToMap(event)
		if marshalErr == nil {
			h.emit(r, internal.Event{Provider: "github", Name: "unknown", Data: data})
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *GitHubHandler) emit(r *http.Request, event internal.Event) {
	topics := h.rules.Evaluate(event)
	for _, topic := range topics {
		if err := h.publisher.Publish(r.Context(), topic, event); err != nil {
			log.Printf("publish %s failed: %v", topic, err)
		}
	}
}

func jsonToMap(payload interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}
