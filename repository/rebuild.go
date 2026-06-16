package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/WindowsSov8forUs/sonolus-core-go/database"
	"github.com/WindowsSov8forUs/sonolus-pack-go/packer"
)

type RebuildOptions struct {
	SourceDir string
	PackDir   string
	TmpDir    string
	Logger    packer.Logger
}

func Rebuild(ctx context.Context, options RebuildOptions) (Snapshot, error) {
	if err := os.MkdirAll(filepath.Join(options.PackDir, "repository"), 0o755); err != nil {
		return Snapshot{}, err
	}

	workDir, err := createPackWorkDir(options.TmpDir, time.Now())
	if err != nil {
		return Snapshot{}, err
	}
	defer os.RemoveAll(workDir)

	if err := packer.Pack(ctx, packer.Options{Input: options.SourceDir, Output: workDir, Logger: options.Logger}); err != nil {
		return Snapshot{}, err
	}
	if err := copyRepository(filepath.Join(workDir, "repository"), filepath.Join(options.PackDir, "repository")); err != nil {
		return Snapshot{}, err
	}

	dbPath := filepath.Join(workDir, "db.json")
	raw, err := os.ReadFile(dbPath)
	if err != nil {
		return Snapshot{}, err
	}
	var db database.Database
	if err := json.Unmarshal(raw, &db); err != nil {
		return Snapshot{}, err
	}

	if err := WriteDatabaseAtomic(filepath.Join(options.PackDir, "db.json"), raw); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Version: time.Now().UnixMilli(), DB: db}, nil
}

func createPackWorkDir(tmpDir string, now time.Time) (string, error) {
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", err
	}
	timestamp := now.UTC().Format("20060102T150405.000000000Z")
	path := filepath.Join(tmpDir, "pack-"+timestamp)
	if err := os.Mkdir(path, 0o755); err == nil {
		return path, nil
	} else if !os.IsExist(err) {
		return "", err
	}
	for attempt := 1; ; attempt++ {
		path = filepath.Join(tmpDir, fmt.Sprintf("pack-%s-%d", timestamp, attempt))
		if err := os.Mkdir(path, 0o755); err == nil {
			return path, nil
		} else if !os.IsExist(err) {
			return "", err
		}
	}
}

func copyRepository(from string, to string) error {
	entries, err := os.ReadDir(from)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := os.MkdirAll(to, 0o755); err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		source := filepath.Join(from, entry.Name())
		target := filepath.Join(to, entry.Name())
		if _, err := os.Stat(target); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return err
		}
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
