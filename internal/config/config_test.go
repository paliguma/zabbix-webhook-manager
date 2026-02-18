package config

import (
	"strings"
	"testing"
)

func TestValidateConfig_AllowsEndpointWithValidAllowedSources(t *testing.T) {
	cfg := Config{
		Port: "8080",
		Endpoints: []Endpoint{
			{
				Name:           "primary",
				Path:           "/webhook/primary",
				AllowedSources: []string{"192.168.1.10", "10.0.0.0/24", "2001:db8::/32"},
			},
		},
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("expected config to be valid, got error: %v", err)
	}
}

func TestValidateConfig_RejectsInvalidAllowedSource(t *testing.T) {
	cfg := Config{
		Port: "8080",
		Endpoints: []Endpoint{
			{
				Name:           "primary",
				Path:           "/webhook/primary",
				AllowedSources: []string{"not-an-ip"},
			},
		},
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Fatalf("expected validation error for invalid allowed source")
	}
	if !strings.Contains(err.Error(), "invalid allowed_sources IP") {
		t.Fatalf("expected invalid allowed_sources IP error, got: %v", err)
	}
}
