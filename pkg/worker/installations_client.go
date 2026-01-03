package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// InstallationRecord mirrors the server installation response.
type InstallationRecord struct {
	Provider       string     `json:"provider"`
	AccountID      string     `json:"account_id"`
	AccountName    string     `json:"account_name"`
	InstallationID string     `json:"installation_id"`
	AccessToken    string     `json:"access_token"`
	RefreshToken   string     `json:"refresh_token"`
	ExpiresAt      *time.Time `json:"expires_at"`
}

// InstallationsClient fetches installation records from the server API.
type InstallationsClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// GetByStateID fetches the latest installation record for a provider/state_id.
func (c *InstallationsClient) GetByStateID(ctx context.Context, provider, stateID string) (*InstallationRecord, error) {
	if stateID == "" {
		return nil, errors.New("state_id is required")
	}
	if provider == "" {
		return nil, errors.New("provider is required")
	}
	base := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if base == "" {
		return nil, errors.New("base url is required")
	}

	endpoint, err := url.Parse(base + "/api/installations")
	if err != nil {
		return nil, err
	}
	q := endpoint.Query()
	q.Set("state_id", stateID)
	q.Set("provider", provider)
	endpoint.RawQuery = q.Encode()

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("installations api failed: %s", resp.Status)
	}
	var records []InstallationRecord
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}
