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

	Token   string `yaml:"token"`
	BaseURL string `yaml:"base_url"`
}
