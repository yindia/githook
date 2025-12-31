package bitbucket

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

// Client is a placeholder Bitbucket client.
type Client struct {
	token   string
	baseURL string
	client  *http.Client
}

// NewTokenClient returns a stub Bitbucket client.
func NewTokenClient(cfg auth.BitbucketConfig, token string) *Client {
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
		return nil, errors.New("bitbucket token is required")
	}
	if owner == "" || repo == "" {
		return nil, errors.New("bitbucket get repo requires workspace and repo")
	}
	endpoint := fmt.Sprintf("%s/repositories/%s/%s", c.baseURL, url.PathEscape(owner), url.PathEscape(repo))
	return c.doJSON(ctx, http.MethodGet, endpoint, nil)
}

func (c *Client) ListRepos(ctx context.Context) ([]map[string]interface{}, error) {
	if c.token == "" {
		return nil, errors.New("bitbucket token is required")
	}
	endpoint := fmt.Sprintf("%s/repositories?role=member", c.baseURL)
	payload, err := c.doJSON(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	raw, ok := payload["values"].([]interface{})
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

func (c *Client) CreatePR(ctx context.Context, owner, repo, title, head, base, body string) (map[string]interface{}, error) {
	if c.token == "" {
		return nil, errors.New("bitbucket token is required")
	}
	if owner == "" || repo == "" {
		return nil, errors.New("bitbucket create pr requires workspace and repo")
	}
	endpoint := fmt.Sprintf("%s/repositories/%s/%s/pullrequests", c.baseURL, url.PathEscape(owner), url.PathEscape(repo))
	input := map[string]interface{}{
		"title": title,
		"source": map[string]interface{}{
			"branch": map[string]interface{}{
				"name": head,
			},
		},
		"destination": map[string]interface{}{
			"branch": map[string]interface{}{
				"name": base,
			},
		},
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
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
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
		return nil, fmt.Errorf("bitbucket api error: %s", string(raw))
	}
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil && err != io.EOF {
		return nil, err
	}
	return out, nil
}

func normalizeBaseURL(base string) string {
	if base == "" {
		return "https://api.bitbucket.org/2.0"
	}
	return strings.TrimRight(base, "/")
}
