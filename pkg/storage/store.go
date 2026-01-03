package storage

import (
	"context"
	"time"
)

// InstallRecord stores SCM installation or token metadata.
type InstallRecord struct {
	Provider       string
	AccountID      string
	AccountName    string
	InstallationID string
	AccessToken    string
	RefreshToken   string
	ExpiresAt      *time.Time
	MetadataJSON   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NamespaceRecord stores provider repository metadata.
type NamespaceRecord struct {
	Provider       string
	AccountID      string
	InstallationID string
	RepoID         string
	Owner          string
	RepoName       string
	FullName       string
	Visibility     string
	DefaultBranch  string
	HTTPURL        string
	SSHURL         string
	WebhooksEnabled bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NamespaceFilter selects namespace rows.
type NamespaceFilter struct {
	Provider  string
	AccountID string
	RepoID    string
	Owner     string
	RepoName  string
	FullName  string
}

// Store defines the persistence interface for installation records.
type Store interface {
	UpsertInstallation(ctx context.Context, record InstallRecord) error
	GetInstallation(ctx context.Context, provider, accountID, installationID string) (*InstallRecord, error)
	GetInstallationByInstallationID(ctx context.Context, provider, installationID string) (*InstallRecord, error)
	ListInstallations(ctx context.Context, provider, accountID string) ([]InstallRecord, error)
	Close() error
}

// NamespaceStore defines persistence for provider repository metadata.
type NamespaceStore interface {
	UpsertNamespace(ctx context.Context, record NamespaceRecord) error
	GetNamespace(ctx context.Context, provider, repoID string) (*NamespaceRecord, error)
	ListNamespaces(ctx context.Context, filter NamespaceFilter) ([]NamespaceRecord, error)
	Close() error
}
