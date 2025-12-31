package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"githooks/internal"

	"github.com/go-playground/webhooks/v6/github"
)

// GitHubHandler handles incoming webhooks from GitHub.
type GitHubHandler struct {
	hook         *github.Webhook
	fallbackHook *github.Webhook
	secret       string
	rules        *internal.RuleEngine
	publisher    internal.Publisher
	logger       *log.Logger
	maxBody      int64
}

var githubEvents = []github.Event{
	github.CheckRunEvent,
	github.CheckSuiteEvent,
	github.CommitCommentEvent,
	github.CreateEvent,
	github.DeleteEvent,
	github.DependabotAlertEvent,
	github.DeployKeyEvent,
	github.DeploymentEvent,
	github.DeploymentStatusEvent,
	github.ForkEvent,
	github.GollumEvent,
	github.InstallationEvent,
	github.InstallationRepositoriesEvent,
	github.IntegrationInstallationEvent,
	github.IntegrationInstallationRepositoriesEvent,
	github.IssueCommentEvent,
	github.IssuesEvent,
	github.LabelEvent,
	github.MemberEvent,
	github.MembershipEvent,
	github.MilestoneEvent,
	github.MetaEvent,
	github.OrganizationEvent,
	github.OrgBlockEvent,
	github.PageBuildEvent,
	github.PingEvent,
	github.ProjectCardEvent,
	github.ProjectColumnEvent,
	github.ProjectEvent,
	github.PublicEvent,
	github.PullRequestEvent,
	github.PullRequestReviewEvent,
	github.PullRequestReviewCommentEvent,
	github.PushEvent,
	github.ReleaseEvent,
	github.RepositoryEvent,
	github.RepositoryVulnerabilityAlertEvent,
	github.SecurityAdvisoryEvent,
	github.StatusEvent,
	github.TeamEvent,
	github.TeamAddEvent,
	github.WatchEvent,
	github.WorkflowDispatchEvent,
	github.WorkflowJobEvent,
	github.WorkflowRunEvent,
	github.GitHubAppAuthorizationEvent,
}

// NewGitHubHandler creates a new GitHubHandler.
func NewGitHubHandler(secret string, rules *internal.RuleEngine, publisher internal.Publisher, logger *log.Logger, maxBody int64) (*GitHubHandler, error) {
	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return nil, err
	}
	fallbackHook, err := github.New()
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = log.Default()
	}
	return &GitHubHandler{
		hook:         hook,
		fallbackHook: fallbackHook,
		secret:       secret,
		rules:        rules,
		publisher:    publisher,
		logger:       logger,
		maxBody:      maxBody,
	}, nil
}

// ServeHTTP handles an incoming HTTP request.
func (h *GitHubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.maxBody > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, h.maxBody)
	}
	reqID := requestID(r)
	w.Header().Set("X-Request-Id", reqID)
	logger := internal.WithRequestID(h.logger, reqID)
	internal.IncRequest("github")
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(rawBody))

	payload, err := h.hook.Parse(r, githubEvents...)
	if err != nil {
		if errors.Is(err, github.ErrMissingHubSignatureHeader) && h.secret != "" {
			sha1Header := r.Header.Get("X-Hub-Signature")
			if sha1Header != "" && verifyGitHubSHA1(h.secret, rawBody, sha1Header) {
				logger.Printf("github parse warning: %v; accepted sha1 signature", err)
				r.Body = io.NopCloser(bytes.NewReader(rawBody))
				payload, err = h.fallbackHook.Parse(r, githubEvents...)
			}
		}
		if err != nil {
			logger.Printf("github parse failed: %v", err)
			internal.IncParseError("github")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	eventName := r.Header.Get("X-GitHub-Event")
	switch payload.(type) {
	case github.PingPayload:
		w.WriteHeader(http.StatusOK)
		return
	default:
		rawObject, data := rawObjectAndFlatten(rawBody)
		h.emit(r, logger, internal.Event{
			Provider:   "github",
			Name:       eventName,
			RequestID:  reqID,
			Data:       data,
			RawPayload: rawBody,
			RawObject:  rawObject,
		})
	}

	w.WriteHeader(http.StatusOK)
}

func (h *GitHubHandler) emit(r *http.Request, logger *log.Logger, event internal.Event) {
	topics := h.rules.EvaluateWithLogger(event, logger)
	logger.Printf("event provider=%s name=%s topics=%v", event.Provider, event.Name, topics)
	for _, match := range topics {
		if err := h.publisher.PublishForDrivers(r.Context(), match.Topic, event, match.Drivers); err != nil {
			logger.Printf("publish %s failed: %v", match.Topic, err)
		}
	}
}

func verifyGitHubSHA1(secret string, body []byte, signature string) bool {
	if secret == "" || len(body) == 0 || signature == "" {
		return false
	}
	signature = strings.TrimPrefix(signature, "sha1=")
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
