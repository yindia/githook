package worker

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Watermill SubscriberConfig `yaml:"watermill"`
}

type RulesConfig struct {
	Rules []struct {
		Emit string `yaml:"emit"`
	} `yaml:"rules"`
}

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
