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
	maxBody   int64
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
func NewGitLabHandler(secret string, rules *internal.RuleEngine, publisher internal.Publisher, logger *log.Logger, maxBody int64) (*GitLabHandler, error) {
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
	return &GitLabHandler{hook: hook, rules: rules, publisher: publisher, logger: logger, maxBody: maxBody}, nil
}

// ServeHTTP handles an incoming HTTP request.
func (h *GitLabHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.maxBody > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, h.maxBody)
	}
	reqID := requestID(r)
	w.Header().Set("X-Request-Id", reqID)
	logger := internal.WithRequestID(h.logger, reqID)
	internal.IncRequest("gitlab")
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(rawBody))

	payload, err := h.hook.Parse(r, gitlabEvents...)
	if err != nil {
		logger.Printf("gitlab parse failed: %v", err)
		internal.IncParseError("gitlab")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventName := r.Header.Get("X-Gitlab-Event")
	switch payload.(type) {
	default:
		rawObject, data := rawObjectAndFlatten(rawBody)
		h.emit(r, logger, internal.Event{
			Provider:   "gitlab",
			Name:       eventName,
			RequestID:  reqID,
			Data:       data,
			RawPayload: rawBody,
			RawObject:  rawObject,
		})
	}

	w.WriteHeader(http.StatusOK)
}

func (h *GitLabHandler) emit(r *http.Request, logger *log.Logger, event internal.Event) {
	topics := h.rules.EvaluateWithLogger(event, logger)
	logger.Printf("event provider=%s name=%s topics=%v", event.Provider, event.Name, topics)
	for _, match := range topics {
		if err := h.publisher.PublishForDrivers(r.Context(), match.Topic, event, match.Drivers); err != nil {
			logger.Printf("publish %s failed: %v", match.Topic, err)
		}
	}
}
