package webhook

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"githooks/internal"

	"github.com/go-playground/webhooks/v6/github"
)

type GitHubHandler struct {
	hook      *github.Webhook
	rules     *internal.RuleEngine
	publisher internal.Publisher
	logger    *log.Logger
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

func NewGitHubHandler(secret string, rules *internal.RuleEngine, publisher internal.Publisher, logger *log.Logger) (*GitHubHandler, error) {
	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = log.Default()
	}
	return &GitHubHandler{hook: hook, rules: rules, publisher: publisher, logger: logger}, nil
}

func (h *GitHubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(rawBody))

	payload, err := h.hook.Parse(r, githubEvents...)
	if err != nil {
		h.logger.Printf("github parse failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventName := r.Header.Get("X-GitHub-Event")
	switch payload.(type) {
	case github.PingPayload:
		w.WriteHeader(http.StatusOK)
		return
	default:
		rawObject, data := rawObjectAndFlatten(rawBody)
		h.emit(r, internal.Event{
			Provider:   "github",
			Name:       eventName,
			Data:       data,
			RawPayload: rawBody,
			RawObject:  rawObject,
		})
	}

	w.WriteHeader(http.StatusOK)
}

func (h *GitHubHandler) emit(r *http.Request, event internal.Event) {
	topics := h.rules.Evaluate(event)
	h.logger.Printf("event provider=%s name=%s topics=%v", event.Provider, event.Name, topics)
	for _, match := range topics {
		if err := h.publisher.PublishForDrivers(r.Context(), match.Topic, event, match.Drivers); err != nil {
			h.logger.Printf("publish %s failed: %v", match.Topic, err)
		}
	}
}

func rawObjectAndFlatten(raw []byte) (interface{}, map[string]interface{}) {
	var out interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, map[string]interface{}{}
	}
	objectMap, ok := out.(map[string]interface{})
	if !ok {
		return out, map[string]interface{}{}
	}
	return out, internal.Flatten(objectMap)
}
