package worker

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// AppConfig is a partial representation of the main application config,
// used for loading worker-specific configuration.
type AppConfig struct {
	Server    ServerConfig    `yaml:"server"`
	Watermill SubscriberConfig `yaml:"watermill"`
}

// ServerConfig is a partial representation of the server config for client resolution.
type ServerConfig struct {
	Port          int    `yaml:"port"`
	PublicBaseURL string `yaml:"public_base_url"`
}

// RulesConfig is a partial representation of the rules configuration,
// used for extracting topic names.
type RulesConfig struct {
	Rules []struct {
		Emit string `yaml:"emit"`
	} `yaml:"rules"`
}

// LoadSubscriberConfig loads the subscriber configuration from a YAML file.
// It expands environment variables and applies default values.
func LoadSubscriberConfig(path string) (SubscriberConfig, error) {
	var cfg AppConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg.Watermill, err
	}
	expanded := os.ExpandEnv(string(data))
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return cfg.Watermill, err
	}
	applySubscriberDefaults(&cfg.Watermill)
	return cfg.Watermill, nil
}

// LoadServerConfig loads the server configuration from a YAML file.
func LoadServerConfig(path string) (ServerConfig, error) {
	var cfg AppConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg.Server, err
	}
	expanded := os.ExpandEnv(string(data))
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return cfg.Server, err
	}
	return cfg.Server, nil
}

// LoadTopicsFromConfig extracts a unique list of topic names from the 'emit' fields
// in a rules configuration file.
func LoadTopicsFromConfig(path string) ([]string, error) {
	var cfg RulesConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	expanded := os.ExpandEnv(string(data))
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}
	topics := make([]string, 0, len(cfg.Rules))
	seen := make(map[string]struct{}, len(cfg.Rules))
	for _, rule := range cfg.Rules {
		topic := strings.TrimSpace(rule.Emit)
		if topic == "" {
			continue
		}
		if _, ok := seen[topic]; ok {
			continue
		}
		seen[topic] = struct{}{}
		topics = append(topics, topic)
	}
	return topics, nil
}

func applySubscriberDefaults(cfg *SubscriberConfig) {
	if cfg.Driver == "" && len(cfg.Drivers) == 0 {
		cfg.Driver = "gochannel"
	}
	if cfg.GoChannel.OutputChannelBuffer == 0 {
		cfg.GoChannel.OutputChannelBuffer = 64
	}
	if cfg.NATS.ClientIDSuffix == "" {
		cfg.NATS.ClientIDSuffix = "-worker"
	}
}
