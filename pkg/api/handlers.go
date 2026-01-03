package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"githooks/pkg/auth"
	"githooks/pkg/oauth"
	"githooks/pkg/storage"
)

// InstallationsHandler lists installations by state/account ID.
type InstallationsHandler struct {
	Store  storage.Store
	Logger *log.Logger
}

func (h *InstallationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.Store == nil {
		http.Error(w, "storage not configured", http.StatusServiceUnavailable)
		return
	}
	accountID := strings.TrimSpace(r.URL.Query().Get("state_id"))
	if accountID == "" {
		accountID = strings.TrimSpace(r.URL.Query().Get("state"))
	}
	if strings.HasPrefix(accountID, "?state_id=") {
		accountID = strings.TrimPrefix(accountID, "?state_id=")
	}
	if accountID == "" {
		http.Error(w, "missing state_id", http.StatusBadRequest)
		return
	}
	provider := strings.TrimSpace(r.URL.Query().Get("provider"))

	var records []storage.InstallRecord
	if provider != "" {
		items, err := h.Store.ListInstallations(r.Context(), provider, accountID)
		if err != nil {
			http.Error(w, "list installations failed", http.StatusInternalServerError)
			if h.Logger != nil {
				h.Logger.Printf("list installations failed: %v", err)
			}
			return
		}
		records = items
	} else {
		providers := []string{"github", "gitlab", "bitbucket"}
		for _, p := range providers {
			items, err := h.Store.ListInstallations(r.Context(), p, accountID)
			if err != nil {
				http.Error(w, "list installations failed", http.StatusInternalServerError)
				if h.Logger != nil {
					h.Logger.Printf("list installations failed: %v", err)
				}
				return
			}
			records = append(records, items...)
		}
	}

	writeJSON(w, records)
}

// NamespacesHandler lists namespaces by filter.
type NamespacesHandler struct {
	Store  storage.NamespaceStore
	Logger *log.Logger
}

func (h *NamespacesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.Store == nil {
		http.Error(w, "storage not configured", http.StatusServiceUnavailable)
		return
	}
	accountID := strings.TrimSpace(r.URL.Query().Get("state_id"))
	if accountID == "" {
		http.Error(w, "missing state_id", http.StatusBadRequest)
		return
	}
	filter := storage.NamespaceFilter{
		Provider:  strings.TrimSpace(r.URL.Query().Get("provider")),
		AccountID: accountID,
		Owner:     strings.TrimSpace(r.URL.Query().Get("owner")),
		RepoName:  strings.TrimSpace(r.URL.Query().Get("repo")),
		FullName:  strings.TrimSpace(r.URL.Query().Get("full_name")),
	}
	records, err := h.Store.ListNamespaces(r.Context(), filter)
	if err != nil {
		http.Error(w, "list namespaces failed", http.StatusInternalServerError)
		if h.Logger != nil {
			h.Logger.Printf("list namespaces failed: %v", err)
		}
		return
	}

	writeJSON(w, records)
}

func writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

// SyncNamespacesHandler triggers a namespace sync for GitLab or Bitbucket.
type SyncNamespacesHandler struct {
	InstallStore  storage.Store
	NamespaceStore storage.NamespaceStore
	Providers     auth.Config
	Logger        *log.Logger
}

func (h *SyncNamespacesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.InstallStore == nil || h.NamespaceStore == nil {
		http.Error(w, "storage not configured", http.StatusServiceUnavailable)
		return
	}
	accountID := strings.TrimSpace(r.URL.Query().Get("state_id"))
	if accountID == "" {
		http.Error(w, "missing state_id", http.StatusBadRequest)
		return
	}
	provider := strings.TrimSpace(r.URL.Query().Get("provider"))
	if provider != "github" && provider != "gitlab" && provider != "bitbucket" {
		http.Error(w, "provider must be github, gitlab, or bitbucket", http.StatusBadRequest)
		return
	}

	record, err := latestInstallation(r.Context(), h.InstallStore, provider, accountID)
	if err != nil {
		http.Error(w, "installation lookup failed", http.StatusInternalServerError)
		if h.Logger != nil {
			h.Logger.Printf("installation lookup failed: %v", err)
		}
		return
	}
	if provider != "github" {
		if record == nil || record.AccessToken == "" {
			http.Error(w, "access token missing", http.StatusBadRequest)
			return
		}
	}

	accessToken := ""
	if record != nil {
		accessToken = record.AccessToken
	}
	if provider != "github" && shouldRefresh(record.ExpiresAt) && record.RefreshToken != "" {
		switch provider {
		case "gitlab":
			refreshed, err := oauth.RefreshGitLabToken(r.Context(), h.Providers.GitLab, record.RefreshToken)
			if err != nil {
				http.Error(w, "token refresh failed", http.StatusInternalServerError)
				if h.Logger != nil {
					h.Logger.Printf("gitlab token refresh failed: %v", err)
				}
				return
			}
			accessToken = refreshed.AccessToken
			record.AccessToken = refreshed.AccessToken
			record.RefreshToken = refreshed.RefreshToken
			record.ExpiresAt = refreshed.ExpiresAt
		case "bitbucket":
			refreshed, err := oauth.RefreshBitbucketToken(r.Context(), h.Providers.Bitbucket, record.RefreshToken)
			if err != nil {
				http.Error(w, "token refresh failed", http.StatusInternalServerError)
				if h.Logger != nil {
					h.Logger.Printf("bitbucket token refresh failed: %v", err)
				}
				return
			}
			accessToken = refreshed.AccessToken
			record.AccessToken = refreshed.AccessToken
			record.RefreshToken = refreshed.RefreshToken
			record.ExpiresAt = refreshed.ExpiresAt
		}
		if err := h.InstallStore.UpsertInstallation(r.Context(), *record); err != nil {
			if h.Logger != nil {
				h.Logger.Printf("token refresh persist failed: %v", err)
			}
		}
	}

	switch provider {
	case "github":
		// No remote sync for GitHub; namespaces come from install webhooks.
	case "gitlab":
		if err := oauth.SyncGitLabNamespaces(r.Context(), h.NamespaceStore, h.Providers.GitLab, accessToken, accountID); err != nil {
			http.Error(w, "namespace sync failed", http.StatusInternalServerError)
			if h.Logger != nil {
				h.Logger.Printf("gitlab namespace sync failed: %v", err)
			}
			return
		}
	case "bitbucket":
		if err := oauth.SyncBitbucketNamespaces(r.Context(), h.NamespaceStore, h.Providers.Bitbucket, accessToken, accountID); err != nil {
			http.Error(w, "namespace sync failed", http.StatusInternalServerError)
			if h.Logger != nil {
				h.Logger.Printf("bitbucket namespace sync failed: %v", err)
			}
			return
		}
	}

	records, err := h.NamespaceStore.ListNamespaces(r.Context(), storage.NamespaceFilter{
		Provider:  provider,
		AccountID: accountID,
	})
	if err != nil {
		http.Error(w, "list namespaces failed", http.StatusInternalServerError)
		if h.Logger != nil {
			h.Logger.Printf("list namespaces failed: %v", err)
		}
		return
	}
	writeJSON(w, records)
}

// NamespaceWebhookHandler toggles webhook enablement for a namespace.
type NamespaceWebhookHandler struct {
	Store          storage.NamespaceStore
	InstallStore   storage.Store
	Providers      auth.Config
	PublicBaseURL  string
	Logger         *log.Logger
}

func (h *NamespaceWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodPost:
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.Store == nil {
		http.Error(w, "storage not configured", http.StatusServiceUnavailable)
		return
	}
	if h.InstallStore == nil {
		http.Error(w, "installation storage not configured", http.StatusServiceUnavailable)
		return
	}
	provider := strings.TrimSpace(r.URL.Query().Get("provider"))
	repoID := strings.TrimSpace(r.URL.Query().Get("repo_id"))
	accountID := strings.TrimSpace(r.URL.Query().Get("state_id"))
	if accountID == "" {
		accountID = strings.TrimSpace(r.URL.Query().Get("state"))
	}
	if provider == "" || repoID == "" || accountID == "" {
		http.Error(w, "missing provider, repo_id, or state_id", http.StatusBadRequest)
		return
	}
	if provider != "github" && provider != "gitlab" && provider != "bitbucket" {
		http.Error(w, "unsupported provider", http.StatusBadRequest)
		return
	}

	record, err := h.Store.GetNamespace(r.Context(), provider, repoID)
	if err != nil {
		http.Error(w, "namespace lookup failed", http.StatusInternalServerError)
		if h.Logger != nil {
			h.Logger.Printf("namespace lookup failed: %v", err)
		}
		return
	}
	if record == nil {
		http.Error(w, "namespace not found", http.StatusNotFound)
		return
	}
	if record.AccountID != accountID {
		http.Error(w, "state_id mismatch", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		if provider == "github" {
			http.Error(w, "github webhooks are always enabled", http.StatusBadRequest)
			return
		}
		webhookURL, err := webhookURL(h.PublicBaseURL, provider)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		install, err := latestInstallation(r.Context(), h.InstallStore, provider, accountID)
		if err != nil || install == nil || install.AccessToken == "" {
			http.Error(w, "access token missing", http.StatusBadRequest)
			return
		}
		enabled := strings.TrimSpace(r.URL.Query().Get("enabled"))
		switch enabled {
		case "true", "1":
			if err := enableProviderWebhook(r.Context(), provider, h.Providers, install.AccessToken, *record, webhookURL); err != nil {
				http.Error(w, "webhook enable failed", http.StatusBadRequest)
				if h.Logger != nil {
					h.Logger.Printf("webhook enable failed: %v", err)
				}
				return
			}
			record.WebhooksEnabled = true
		case "false", "0":
			if err := disableProviderWebhook(r.Context(), provider, h.Providers, install.AccessToken, *record, webhookURL); err != nil {
				http.Error(w, "webhook disable failed", http.StatusBadRequest)
				if h.Logger != nil {
					h.Logger.Printf("webhook disable failed: %v", err)
				}
				return
			}
			record.WebhooksEnabled = false
		default:
			http.Error(w, "enabled must be true or false", http.StatusBadRequest)
			return
		}
		if err := h.Store.UpsertNamespace(r.Context(), *record); err != nil {
			http.Error(w, "namespace update failed", http.StatusInternalServerError)
			if h.Logger != nil {
				h.Logger.Printf("namespace update failed: %v", err)
			}
			return
		}
	}

	writeJSON(w, map[string]bool{
		"enabled": record.WebhooksEnabled,
	})
}
func latestInstallation(ctx context.Context, store storage.Store, provider, accountID string) (*storage.InstallRecord, error) {
	records, err := store.ListInstallations(ctx, provider, accountID)
	if err != nil {
		return nil, err
	}
	var latest *storage.InstallRecord
	for i := range records {
		item := records[i]
		if latest == nil || item.UpdatedAt.After(latest.UpdatedAt) {
			copy := item
			latest = &copy
		}
	}
	return latest, nil
}

func shouldRefresh(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return time.Now().UTC().After(expiresAt.Add(-1 * time.Minute))
}

func webhookURL(publicBaseURL, provider string) (string, error) {
	publicBaseURL = strings.TrimSpace(publicBaseURL)
	publicBaseURL = strings.TrimRight(publicBaseURL, "/")
	if publicBaseURL == "" {
		return "", errors.New("public_base_url is required for webhook management")
	}
	switch provider {
	case "gitlab":
		return publicBaseURL + "/webhooks/gitlab", nil
	case "bitbucket":
		return publicBaseURL + "/webhooks/bitbucket", nil
	default:
		return "", errors.New("unsupported provider for webhook management")
	}
}
