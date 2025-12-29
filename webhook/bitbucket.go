package webhook

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"githooks/internal"

	"github.com/go-playground/webhooks/v6/bitbucket"
)

type BitbucketHandler struct {
	hook      *bitbucket.Webhook
	rules     *internal.RuleEngine
	publisher internal.Publisher
	logger    *log.Logger
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

func NewBitbucketHandler(secret string, rules *internal.RuleEngine, publisher internal.Publisher, logger *log.Logger) (*BitbucketHandler, error) {
	hook, err := bitbucket.New(bitbucket.Options.UUID(secret))
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = log.Default()
	}
	return &BitbucketHandler{hook: hook, rules: rules, publisher: publisher, logger: logger}, nil
}

func (h *BitbucketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(rawBody))

	payload, err := h.hook.Parse(r, bitbucketEvents...)
	if err != nil {
		h.logger.Printf("bitbucket parse failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventName := r.Header.Get("X-Event-Key")
	switch payload.(type) {
	default:
		rawObject, data := rawObjectAndFlatten(rawBody)
		h.emit(r, internal.Event{
			Provider:   "bitbucket",
			Name:       eventName,
			Data:       data,
			RawPayload: rawBody,
			RawObject:  rawObject,
		})
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BitbucketHandler) emit(r *http.Request, event internal.Event) {
	topics := h.rules.Evaluate(event)
	h.logger.Printf("event provider=%s name=%s topics=%v", event.Provider, event.Name, topics)
	for _, match := range topics {
		if err := h.publisher.PublishForDrivers(r.Context(), match.Topic, event, match.Drivers); err != nil {
			h.logger.Printf("publish %s failed: %v", match.Topic, err)
		}
	}
}
