package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"githooks/pkg/auth"
)

// Client is a placeholder GitLab client.
type Client struct {
	token   string
	baseURL string
	client  *http.Client
}

// NewTokenClient returns a stub GitLab client.
func NewTokenClient(cfg auth.ProviderConfig, token string) *Client {
	if token == "" {
		token = cfg.Token
	}
	return &Client{
		token:   token,
		baseURL: normalizeBaseURL(cfg.BaseURL),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) GetRepo(ctx context.Context, owner, repo string) (map[string]interface{}, error) {
	if c.token == "" {
		return nil, errors.New("gitlab token is required")
	}
	if owner == "" || repo == "" {
		return nil, errors.New("gitlab get repo requires owner and repo")
	}
	project := url.PathEscape(fmt.Sprintf("%s/%s", owner, repo))
	endpoint := fmt.Sprintf("%s/projects/%s", c.baseURL, project)
	return c.doJSON(ctx, http.MethodGet, endpoint, nil)
}

func (c *Client) ListRepos(ctx context.Context) ([]map[string]interface{}, error) {
	if c.token == "" {
		return nil, errors.New("gitlab token is required")
	}
	endpoint := fmt.Sprintf("%s/projects?membership=true", c.baseURL)
	return c.doJSONArray(ctx, http.MethodGet, endpoint)
}

func (c *Client) CreatePR(ctx context.Context, owner, repo, title, head, base, body string) (map[string]interface{}, error) {
	if c.token == "" {
		return nil, errors.New("gitlab token is required")
	}
	if owner == "" || repo == "" {
		return nil, errors.New("gitlab create pr requires owner and repo")
	}
	project := url.PathEscape(fmt.Sprintf("%s/%s", owner, repo))
	endpoint := fmt.Sprintf("%s/projects/%s/merge_requests", c.baseURL, project)
	input := map[string]interface{}{
		"title":         title,
		"source_branch": head,
		"target_branch": base,
	}
	if body != "" {
		input["description"] = body
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
	req.Header.Set("PRIVATE-TOKEN", c.token)
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
		return nil, fmt.Errorf("gitlab api error: %s", string(raw))
	}
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil && err != io.EOF {
		return nil, err
	}
	return out, nil
}

func (c *Client) doJSONArray(ctx context.Context, method, url string) ([]map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitlab api error: %s", string(raw))
	}
	var out []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil && err != io.EOF {
		return nil, err
	}
	return out, nil
}

func normalizeBaseURL(base string) string {
	if base == "" {
		return "https://gitlab.com/api/v4"
	}
	return strings.TrimRight(base, "/")
}
