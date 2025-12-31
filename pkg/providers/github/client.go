package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is a minimal GitHub API client using an installation token.
type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewAppClient creates a GitHub client by exchanging an installation token.
func NewAppClient(ctx context.Context, cfg AppConfig, installationID int64) (*Client, error) {
	if installationID == 0 {
		return nil, fmt.Errorf("github installation id is required")
	}
	authenticator := newAppAuthenticator(cfg)
	token, err := authenticator.installationToken(ctx, installationID)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL: normalizeBaseURL(cfg.BaseURL),
		token:   token,
		client:  &http.Client{},
	}, nil
}

// GetRepo fetches repository metadata.
func (c *Client) GetRepo(ctx context.Context, owner, repo string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)
	return c.doJSON(ctx, http.MethodGet, endpoint, nil)
}

// ListRepos lists repositories accessible to the installation.
func (c *Client) ListRepos(ctx context.Context) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/installation/repositories", c.baseURL)
	payload, err := c.doJSON(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	raw, ok := payload["repositories"].([]interface{})
	if !ok {
		return nil, nil
	}
	repos := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		if repo, ok := item.(map[string]interface{}); ok {
			repos = append(repos, repo)
		}
	}
	return repos, nil
}

// CreatePR opens a pull request in the target repository.
func (c *Client) CreatePR(ctx context.Context, owner, repo, title, head, base, body string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/pulls", c.baseURL, owner, repo)
	input := map[string]interface{}{
		"title": title,
		"head":  head,
		"base":  base,
	}
	if body != "" {
		input["body"] = body
	}
	return c.doJSON(ctx, http.MethodPost, endpoint, input)
}

func (c *Client) doJSON(ctx context.Context, method, url string, payload map[string]interface{}) (map[string]interface{}, error) {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error: %s", string(raw))
	}
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil && err != io.EOF {
		return nil, err
	}
	return out, nil
}
