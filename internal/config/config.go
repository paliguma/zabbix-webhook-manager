package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultPort         = "8080"
	defaultEndpointName = "default"
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
	Port         string     `yaml:"port"`
	WebhookPath  string     `yaml:"webhook_path"`
	Endpoints    []Endpoint `yaml:"endpoints"`
	LogHeaders   bool       `yaml:"log_headers"`
	LogBody      bool       `yaml:"log_body"`
	MaxBodyBytes int64      `yaml:"max_body_bytes"`
}

type Endpoint struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

func Load() (Config, string, bool, error) {
	cfg := defaultConfig()
	path := configPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			applyDefaults(&cfg)
			applyEnvOverrides(&cfg)
			if err := validateConfig(cfg); err != nil {
				return Config{}, path, false, err
			}
			return cfg, path, false, nil
		}
		return Config{}, path, false, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, path, true, err
	}

	applyDefaults(&cfg)
	applyEnvOverrides(&cfg)
	if err := validateConfig(cfg); err != nil {
		return Config{}, path, true, err
	}
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
		Port:        defaultPort,
		WebhookPath: defaultWebhookPath,
		Endpoints: []Endpoint{
			{
				Name: defaultEndpointName,
				Path: defaultWebhookPath,
			},
		},
		LogHeaders:   true,
		LogBody:      true,
		MaxBodyBytes: defaultMaxBodyBytes,
	}
}

func applyDefaults(cfg *Config) {
	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = defaultMaxBodyBytes
	}
	normalizeEndpoints(cfg)
	if cfg.WebhookPath == "" {
		cfg.WebhookPath = cfg.Endpoints[0].Path
	}
}

func applyEnvOverrides(cfg *Config) {
	if envPort := os.Getenv("PORT"); envPort != "" {
		cfg.Port = envPort
	}
	if envPath := os.Getenv("WEBHOOK_PATH"); envPath != "" {
		cfg.WebhookPath = envPath
		cfg.Endpoints = []Endpoint{
			{
				Name: defaultEndpointName,
				Path: envPath,
			},
		}
	}
}

func normalizeEndpoints(cfg *Config) {
	endpoints := make([]Endpoint, 0, len(cfg.Endpoints))
	for _, endpoint := range cfg.Endpoints {
		if endpoint.Path == "" {
			continue
		}
		if endpoint.Name == "" {
			endpoint.Name = endpoint.Path
		}
		endpoints = append(endpoints, endpoint)
	}

	if len(endpoints) == 0 {
		path := cfg.WebhookPath
		if path == "" {
			path = defaultWebhookPath
		}
		endpoints = append(endpoints, Endpoint{
			Name: defaultEndpointName,
			Path: path,
		})
	}

	cfg.Endpoints = endpoints
}

func validateConfig(cfg Config) error {
	if len(cfg.Endpoints) == 0 {
		return errors.New("at least one webhook endpoint must be configured")
	}

	seenPaths := map[string]string{}
	for _, endpoint := range cfg.Endpoints {
		if endpoint.Path == "" {
			return errors.New("endpoint path cannot be empty")
		}
		if !strings.HasPrefix(endpoint.Path, "/") {
			return fmt.Errorf("endpoint path %q must start with '/'", endpoint.Path)
		}
		if existing, exists := seenPaths[endpoint.Path]; exists {
			return fmt.Errorf(
				"duplicate endpoint path %q configured for %q and %q",
				endpoint.Path,
				existing,
				endpoint.Name,
			)
		}
		seenPaths[endpoint.Path] = endpoint.Name
	}

	return nil
}
