package webhook

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"githooks/internal"

	"github.com/go-playground/webhooks/v6/gitlab"
)

// GitLabHandler handles incoming webhooks from GitLab.
type GitLabHandler struct {
	hook      *gitlab.Webhook
	rules     *internal.RuleEngine
	publisher internal.Publisher
	logger    *log.Logger
}

var gitlabEvents = []gitlab.Event{
	gitlab.PushEvents,
	gitlab.TagEvents,
	gitlab.IssuesEvents,
	gitlab.ConfidentialIssuesEvents,
	gitlab.CommentEvents,
	gitlab.ConfidentialCommentEvents,
	gitlab.MergeRequestEvents,
	gitlab.WikiPageEvents,
	gitlab.PipelineEvents,
	gitlab.BuildEvents,
	gitlab.JobEvents,
	gitlab.DeploymentEvents,
	gitlab.SystemHookEvents,
}

// NewGitLabHandler creates a new GitLabHandler.
func NewGitLabHandler(secret string, rules *internal.RuleEngine, publisher internal.Publisher, logger *log.Logger) (*GitLabHandler, error) {
	options := make([]gitlab.Option, 0, 1)
	if secret != "" {
		options = append(options, gitlab.Options.Secret(secret))
	}
	hook, err := gitlab.New(options...)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = log.Default()
	}
	return &GitLabHandler{hook: hook, rules: rules, publisher: publisher, logger: logger}, nil
}

// ServeHTTP handles an incoming HTTP request.
func (h *GitLabHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(rawBody))

	payload, err := h.hook.Parse(r, gitlabEvents...)
	if err != nil {
		h.logger.Printf("gitlab parse failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventName := r.Header.Get("X-Gitlab-Event")
	switch payload.(type) {
	default:
		rawObject, data := rawObjectAndFlatten(rawBody)
		h.emit(r, internal.Event{
			Provider:   "gitlab",
			Name:       eventName,
			Data:       data,
			RawPayload: rawBody,
			RawObject:  rawObject,
		})
	}

	w.WriteHeader(http.StatusOK)
}

func (h *GitLabHandler) emit(r *http.Request, event internal.Event) {
	topics := h.rules.Evaluate(event)
	h.logger.Printf("event provider=%s name=%s topics=%v", event.Provider, event.Name, topics)
	for _, match := range topics {
		if err := h.publisher.PublishForDrivers(r.Context(), match.Topic, event, match.Drivers); err != nil {
			h.logger.Printf("publish %s failed: %v", match.Topic, err)
		}
	}
}
