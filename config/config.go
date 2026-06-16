package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/ini.v1"
)

type Config struct {
	Server     ServerConfig
	Repository RepositoryConfig
}

type ServerConfig struct {
	Listen string
}

type RepositoryConfig struct {
	SourceDir    string
	PackDir      string
	TmpDir       string
	WatchSource  bool
	PollInterval time.Duration
}

func Load(path string) Config {
	cfg := Config{
		Server: ServerConfig{
			Listen: "127.0.0.1:8000",
		},
		Repository: RepositoryConfig{
			SourceDir:    abs("source"),
			PackDir:      abs("pack"),
			TmpDir:       abs("tmp"),
			WatchSource:  false,
			PollInterval: 10 * time.Second,
		},
	}
	if path != "" {
		if file, err := ini.Load(path); err == nil {
			server := file.Section("server")
			if listen := server.Key("listen").String(); listen != "" {
				cfg.Server.Listen = listen
			} else if port := server.Key("port").String(); port != "" {
				cfg.Server.Listen = "127.0.0.1:" + port
			}

			repository := file.Section("repository")
			cfg.Repository.SourceDir = abs(repository.Key("source-dir").MustString(cfg.Repository.SourceDir))
			cfg.Repository.PackDir = abs(repository.Key("pack-dir").MustString(cfg.Repository.PackDir))
			cfg.Repository.TmpDir = abs(repository.Key("tmp-dir").MustString(cfg.Repository.TmpDir))
			cfg.Repository.WatchSource = repository.Key("watch-source").MustBool(cfg.Repository.WatchSource)
			if value := repository.Key("poll-interval").String(); value != "" {
				if duration, err := time.ParseDuration(value); err == nil {
					cfg.Repository.PollInterval = duration
				}
			}
		}
	}
	cfg.Repository.SourceDir = abs(envOrDefault("SONOLUS_REPOSITORY_SOURCE_DIR", cfg.Repository.SourceDir))
	cfg.Repository.PackDir = abs(envOrDefault("SONOLUS_REPOSITORY_PACK_DIR", cfg.Repository.PackDir))
	cfg.Repository.TmpDir = abs(envOrDefault("SONOLUS_REPOSITORY_TMP_DIR", cfg.Repository.TmpDir))
	if value := os.Getenv("SONOLUS_REPOSITORY_WATCH_SOURCE"); value != "" {
		cfg.Repository.WatchSource = value != "0"
	}
	if value := os.Getenv("SONOLUS_REPOSITORY_POLL_INTERVAL"); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			cfg.Repository.PollInterval = duration
		}
	}
	if listen := os.Getenv("SONOLUS_LISTEN_ADDR"); listen != "" {
		cfg.Server.Listen = listen
	} else if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Listen = "127.0.0.1:" + port
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
