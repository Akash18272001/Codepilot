package config

import (
	"fmt"
	"os"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/providers/env"
)

type Config struct {
	Port        string
	OllamaHost  string
	OllamaModel string
	GitHubToken string
	GitHubOwner string
	GitHubRepo  string
}

var k = koanf.New(".")

func LoadConfig() (*Config, error) {
	_ = k.Load(env.Provider("", ".", func(s string) string {
		return s
	}), nil)

	cfg := &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		OllamaHost:  os.Getenv("OLLAMA_HOST"),
		OllamaModel: os.Getenv("OLLAMA_MODEL"),
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		GitHubOwner: os.Getenv("GITHUB_OWNER"),
		GitHubRepo:  os.Getenv("GITHUB_REPO"),
	}

	// 🚨 Production Fail-Fast Validation (No silent defaults for critical secrets)
	if cfg.OllamaHost == "" {
		return nil, fmt.Errorf("critical config missing: OLLAMA_HOST is required")
	}
	if cfg.OllamaModel == "" {
		return nil, fmt.Errorf("critical config missing: OLLAMA_MODEL is required")
	}
	if cfg.GitHubToken == "" {
		return nil, fmt.Errorf("critical config missing: GITHUB_TOKEN is required")
	}
	if cfg.GitHubOwner == "" {
		return nil, fmt.Errorf("critical config missing: GITHUB_OWNER is required")
	}
	if cfg.GitHubRepo == "" {
		return nil, fmt.Errorf("critical config missing: GITHUB_REPO is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}