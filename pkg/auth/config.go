package auth

// Config contains provider configuration for webhooks and SCM auth.
type Config struct {
	GitHub    ProviderConfig `yaml:"github"`
	GitLab    ProviderConfig `yaml:"gitlab"`
	Bitbucket ProviderConfig `yaml:"bitbucket"`
}

// ProviderConfig contains webhook and auth configuration for a provider.
type ProviderConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
	Secret  string `yaml:"secret"`

	AppID          int64  `yaml:"app_id"`
	PrivateKeyPath string `yaml:"private_key_path"`
	AppSlug        string `yaml:"app_slug"`
	WebBaseURL     string `yaml:"web_base_url"`

	BaseURL string `yaml:"base_url"`

	OAuthClientID     string   `yaml:"oauth_client_id"`
	OAuthClientSecret string   `yaml:"oauth_client_secret"`
	OAuthScopes       []string `yaml:"oauth_scopes"`
}
