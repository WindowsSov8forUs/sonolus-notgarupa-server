package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadReadsSectionedConfig(t *testing.T) {
	clearEnv(t)
	root := t.TempDir()
	source := filepath.Join(root, "source")
	pack := filepath.Join(root, "pack")
	tmp := filepath.Join(root, "tmp")
	path := filepath.Join(root, "config.ini")
	data := []byte("[server]\nport = 9000\n\n[repository]\nsource-dir = " + source + "\npack-dir = " + pack + "\ntmp-dir = " + tmp + "\nwatch-source = false\npoll-interval = 2s\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := Load(path)

	if cfg.Server.Listen != "127.0.0.1:9000" {
		t.Fatalf("listen=%q", cfg.Server.Listen)
	}
	if cfg.Repository.SourceDir != source {
		t.Fatalf("source=%q, want %q", cfg.Repository.SourceDir, source)
	}
	if cfg.Repository.PackDir != pack {
		t.Fatalf("pack=%q, want %q", cfg.Repository.PackDir, pack)
	}
	if cfg.Repository.TmpDir != tmp {
		t.Fatalf("tmp=%q, want %q", cfg.Repository.TmpDir, tmp)
	}
	if cfg.Repository.WatchSource {
		t.Fatal("watch source=true, want false")
	}
	if cfg.Repository.PollInterval != 2*time.Second {
		t.Fatalf("poll interval=%s, want 2s", cfg.Repository.PollInterval)
	}
}

func TestLoadDefaultsWatchSourceToFalse(t *testing.T) {
	clearEnv(t)
	cfg := Load("")

	if cfg.Repository.WatchSource {
		t.Fatal("watch source=true, want false")
	}
}

func TestLoadEnvironmentOverridesSectionedConfig(t *testing.T) {
	clearEnv(t)
	root := t.TempDir()
	path := filepath.Join(root, "config.ini")
	if err := os.WriteFile(path, []byte("[server]\nport = 9000\n\n[repository]\nwatch-source = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "env-source")
	pack := filepath.Join(root, "env-pack")
	tmp := filepath.Join(root, "env-tmp")
	t.Setenv("SONOLUS_LISTEN_ADDR", "0.0.0.0:7000")
	t.Setenv("SONOLUS_REPOSITORY_SOURCE_DIR", source)
	t.Setenv("SONOLUS_REPOSITORY_PACK_DIR", pack)
	t.Setenv("SONOLUS_REPOSITORY_TMP_DIR", tmp)
	t.Setenv("SONOLUS_REPOSITORY_WATCH_SOURCE", "0")
	t.Setenv("SONOLUS_REPOSITORY_POLL_INTERVAL", "5s")

	cfg := Load(path)

	if cfg.Server.Listen != "0.0.0.0:7000" {
		t.Fatalf("listen=%q", cfg.Server.Listen)
	}
	if cfg.Repository.SourceDir != source {
		t.Fatalf("source=%q, want %q", cfg.Repository.SourceDir, source)
	}
	if cfg.Repository.PackDir != pack {
		t.Fatalf("pack=%q, want %q", cfg.Repository.PackDir, pack)
	}
	if cfg.Repository.TmpDir != tmp {
		t.Fatalf("tmp=%q, want %q", cfg.Repository.TmpDir, tmp)
	}
	if cfg.Repository.WatchSource {
		t.Fatal("watch source=true, want false")
	}
	if cfg.Repository.PollInterval != 5*time.Second {
		t.Fatalf("poll interval=%s, want 5s", cfg.Repository.PollInterval)
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"PORT",
		"SONOLUS_LISTEN_ADDR",
		"SONOLUS_REPOSITORY_SOURCE_DIR",
		"SONOLUS_REPOSITORY_PACK_DIR",
		"SONOLUS_REPOSITORY_TMP_DIR",
		"SONOLUS_REPOSITORY_WATCH_SOURCE",
		"SONOLUS_REPOSITORY_POLL_INTERVAL",
	} {
		t.Setenv(name, "")
	}
}
