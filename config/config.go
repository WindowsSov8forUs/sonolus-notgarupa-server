package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/ini.v1"
)

type Config struct {
	Address string
	Listen  string

	RepositoryPublicURL    string
	RepositorySourceDir    string
	RepositoryDataDir      string
	RepositoryTmpDir       string
	RepositoryWatchSource  bool
	RepositoryPollInterval time.Duration
}

func Load(path string) Config {
	cfg := Config{
		Address: "http://localhost:8000",
		Listen:  "127.0.0.1:8000",

		RepositoryPublicURL:    "http://localhost:8000",
		RepositorySourceDir:    abs("source"),
		RepositoryDataDir:      abs("data"),
		RepositoryTmpDir:       abs("tmp"),
		RepositoryWatchSource:  os.Getenv("SONOLUS_REPOSITORY_WATCH_SOURCE") != "0",
		RepositoryPollInterval: 10 * time.Second,
	}
	if path != "" {
		if file, err := ini.Load(path); err == nil {
			section := file.Section("")
			cfg.Address = section.Key("address").MustString(cfg.Address)
			cfg.RepositoryPublicURL = section.Key("repository-public-url").MustString(cfg.RepositoryPublicURL)
			cfg.RepositorySourceDir = abs(section.Key("repository-source-dir").MustString(cfg.RepositorySourceDir))
			cfg.RepositoryDataDir = abs(section.Key("repository-data-dir").MustString(cfg.RepositoryDataDir))
			cfg.RepositoryTmpDir = abs(section.Key("repository-tmp-dir").MustString(cfg.RepositoryTmpDir))
			cfg.RepositoryWatchSource = section.Key("repository-watch-source").MustBool(cfg.RepositoryWatchSource)
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
	cfg.RepositoryPublicURL = envOrDefault("SONOLUS_REPOSITORY_PUBLIC_URL", cfg.RepositoryPublicURL)
	cfg.RepositorySourceDir = abs(envOrDefault("SONOLUS_REPOSITORY_SOURCE_DIR", cfg.RepositorySourceDir))
	cfg.RepositoryDataDir = abs(envOrDefault("SONOLUS_REPOSITORY_DATA_DIR", cfg.RepositoryDataDir))
	cfg.RepositoryTmpDir = abs(envOrDefault("SONOLUS_REPOSITORY_TMP_DIR", cfg.RepositoryTmpDir))
	if value := os.Getenv("SONOLUS_REPOSITORY_WATCH_SOURCE"); value != "" {
		cfg.RepositoryWatchSource = value != "0"
	}
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
	if cfg.RepositoryPublicURL == "" {
		cfg.RepositoryPublicURL = cfg.Address
	}
	return cfg
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func abs(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return absolute
}
