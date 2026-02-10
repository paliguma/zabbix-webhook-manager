package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	defaultPort         = "8080"
	defaultWebhookPath  = "/webhook"
	defaultMaxBodyBytes = int64(1 << 20)
	defaultConfigPath   = "configs/config.yml"
)

var candidateConfigPaths = []string{
	"configs/config.yml",
	"configs/config.yaml",
	"configs/config.json",
	"config.yml",
	"config.yaml",
	"config.json",
}

type Config struct {
	Port         string `yaml:"port"`
	WebhookPath  string `yaml:"webhook_path"`
	LogHeaders   bool   `yaml:"log_headers"`
	LogBody      bool   `yaml:"log_body"`
	MaxBodyBytes int64  `yaml:"max_body_bytes"`
}

func Load() (Config, string, bool, error) {
	cfg := defaultConfig()
	path := configPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			applyEnvOverrides(&cfg)
			return cfg, path, false, nil
		}
		return Config{}, path, false, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, path, true, err
	}

	applyDefaults(&cfg)
	applyEnvOverrides(&cfg)
	return cfg, path, true, nil
}

func configPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	for _, path := range candidateConfigPaths {
		if fileExists(path) {
			return path
		}
	}
	return defaultConfigPath
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func defaultConfig() Config {
	return Config{
		Port:         defaultPort,
		WebhookPath:  defaultWebhookPath,
		LogHeaders:   true,
		LogBody:      true,
		MaxBodyBytes: defaultMaxBodyBytes,
	}
}

func applyDefaults(cfg *Config) {
	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	if cfg.WebhookPath == "" {
		cfg.WebhookPath = defaultWebhookPath
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = defaultMaxBodyBytes
	}
}

func applyEnvOverrides(cfg *Config) {
	if envPort := os.Getenv("PORT"); envPort != "" {
		cfg.Port = envPort
	}
	if envPath := os.Getenv("WEBHOOK_PATH"); envPath != "" {
		cfg.WebhookPath = envPath
	}
}
