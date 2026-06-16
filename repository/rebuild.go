package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/WindowsSov8forUs/sonolus-core-go/database"
	"github.com/WindowsSov8forUs/sonolus-pack-go/packer"
)

type RebuildOptions struct {
	SourceDir string
	DataDir   string
	TmpDir    string
	PublicURL string
	Logger    packer.Logger
}

func Rebuild(ctx context.Context, options RebuildOptions) (Manifest, error) {
	if err := os.MkdirAll(options.TmpDir, 0o755); err != nil {
		return Manifest{}, err
	}
	if err := os.MkdirAll(filepath.Join(options.DataDir, "repository"), 0o755); err != nil {
		return Manifest{}, err
	}

	workDir, err := os.MkdirTemp(options.TmpDir, "pack-*")
	if err != nil {
		return Manifest{}, err
	}
	defer os.RemoveAll(workDir)

	if err := packer.Pack(ctx, packer.Options{Input: options.SourceDir, Output: workDir, Logger: options.Logger}); err != nil {
		return Manifest{}, err
	}
	if err := copyRepository(filepath.Join(workDir, "repository"), filepath.Join(options.DataDir, "repository")); err != nil {
		return Manifest{}, err
	}

	dbPath := filepath.Join(workDir, "db.json")
	raw, err := os.ReadFile(dbPath)
	if err != nil {
		return Manifest{}, err
	}
	var dbValue any
	if err := json.Unmarshal(raw, &dbValue); err != nil {
		return Manifest{}, err
	}
	RewriteRepositoryURLs(dbValue, options.PublicURL)
	rewritten, err := json.Marshal(dbValue)
	if err != nil {
		return Manifest{}, err
	}

	var db database.Database
	if err := json.Unmarshal(rewritten, &db); err != nil {
		return Manifest{}, err
	}

	now := time.Now()
	manifest := Manifest{
		Version:     now.UnixMilli(),
		GeneratedAt: now.UTC().Format(time.RFC3339Nano),
		DB:          db,
	}
	if err := WriteManifestAtomic(filepath.Join(options.DataDir, "manifest.json"), manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
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
