package config

import (
	"os"
	"time"

	"gopkg.in/ini.v1"
)

type Config struct {
	Address string
	Listen  string

	RepositoryAdminURL     string
	RepositoryManifestURL  string
	RepositoryPollInterval time.Duration
}

func Load(path string) Config {
	cfg := Config{
		Address: "http://localhost:8000",
		Listen:  "127.0.0.1:8000",

		RepositoryAdminURL:     "http://127.0.0.1:9000",
		RepositoryManifestURL:  "http://localhost:9000/manifest.json",
		RepositoryPollInterval: 10 * time.Second,
	}
	if path != "" {
		if file, err := ini.Load(path); err == nil {
			section := file.Section("")
			cfg.Address = section.Key("address").MustString(cfg.Address)
			cfg.RepositoryAdminURL = section.Key("repository-admin-url").MustString(cfg.RepositoryAdminURL)
			cfg.RepositoryManifestURL = section.Key("repository-manifest-url").MustString(cfg.RepositoryManifestURL)
			if value := section.Key("repository-poll-interval").String(); value != "" {
				if duration, err := time.ParseDuration(value); err == nil {
					cfg.RepositoryPollInterval = duration
				}
			}
			if port := section.Key("port").String(); port != "" {
				cfg.Listen = "127.0.0.1:" + port
			}
		}
	}
	cfg.Address = envOrDefault("SONOLUS_ADDRESS", cfg.Address)
	cfg.RepositoryAdminURL = envOrDefault("SONOLUS_REPOSITORY_ADMIN_URL", cfg.RepositoryAdminURL)
	cfg.RepositoryManifestURL = envOrDefault("SONOLUS_REPOSITORY_MANIFEST_URL", cfg.RepositoryManifestURL)
	if value := os.Getenv("SONOLUS_REPOSITORY_POLL_INTERVAL"); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			cfg.RepositoryPollInterval = duration
		}
	}
	if listen := os.Getenv("SONOLUS_LISTEN_ADDR"); listen != "" {
		cfg.Listen = listen
	} else if port := os.Getenv("PORT"); port != "" {
		cfg.Listen = "127.0.0.1:" + port
	}
	return cfg
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
