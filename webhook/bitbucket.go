package webhook

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"githooks/internal"

	"github.com/go-playground/webhooks/v6/bitbucket"
)

// BitbucketHandler handles incoming webhooks from Bitbucket.
type BitbucketHandler struct {
	hook      *bitbucket.Webhook
	rules     *internal.RuleEngine
	publisher internal.Publisher
	logger    *log.Logger
	maxBody   int64
}

var bitbucketEvents = []bitbucket.Event{
	bitbucket.RepoPushEvent,
	bitbucket.RepoForkEvent,
	bitbucket.RepoUpdatedEvent,
	bitbucket.RepoCommitCommentCreatedEvent,
	bitbucket.RepoCommitStatusCreatedEvent,
	bitbucket.RepoCommitStatusUpdatedEvent,
	bitbucket.IssueCreatedEvent,
	bitbucket.IssueUpdatedEvent,
	bitbucket.IssueCommentCreatedEvent,
	bitbucket.PullRequestCreatedEvent,
	bitbucket.PullRequestUpdatedEvent,
	bitbucket.PullRequestApprovedEvent,
	bitbucket.PullRequestUnapprovedEvent,
	bitbucket.PullRequestMergedEvent,
	bitbucket.PullRequestDeclinedEvent,
	bitbucket.PullRequestCommentCreatedEvent,
	bitbucket.PullRequestCommentUpdatedEvent,
	bitbucket.PullRequestCommentDeletedEvent,
}

// NewBitbucketHandler creates a new BitbucketHandler.
func NewBitbucketHandler(secret string, rules *internal.RuleEngine, publisher internal.Publisher, logger *log.Logger, maxBody int64) (*BitbucketHandler, error) {
	options := make([]bitbucket.Option, 0, 1)
	if secret != "" {
		options = append(options, bitbucket.Options.UUID(secret))
	}
	hook, err := bitbucket.New(options...)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = log.Default()
	}
	return &BitbucketHandler{hook: hook, rules: rules, publisher: publisher, logger: logger, maxBody: maxBody}, nil
}

// ServeHTTP handles an incoming HTTP request.
func (h *BitbucketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.maxBody > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, h.maxBody)
	}
	reqID := requestID(r)
	w.Header().Set("X-Request-Id", reqID)
	logger := internal.WithRequestID(h.logger, reqID)
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(rawBody))

	payload, err := h.hook.Parse(r, bitbucketEvents...)
	if err != nil {
		logger.Printf("bitbucket parse failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventName := r.Header.Get("X-Event-Key")
	switch payload.(type) {
	default:
		rawObject, data := rawObjectAndFlatten(rawBody)
		h.emit(r, logger, internal.Event{
			Provider:   "bitbucket",
			Name:       eventName,
			Data:       data,
			RawPayload: rawBody,
			RawObject:  rawObject,
		})
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BitbucketHandler) emit(r *http.Request, logger *log.Logger, event internal.Event) {
	topics := h.rules.Evaluate(event)
	logger.Printf("event provider=%s name=%s topics=%v", event.Provider, event.Name, topics)
	for _, match := range topics {
		if err := h.publisher.PublishForDrivers(r.Context(), match.Topic, event, match.Drivers); err != nil {
			logger.Printf("publish %s failed: %v", match.Topic, err)
		}
	}
}
