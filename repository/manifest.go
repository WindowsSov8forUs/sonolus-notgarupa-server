package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/WindowsSov8forUs/sonolus-core-go/database"
)

type Manifest struct {
	Version     int64             `json:"version"`
	GeneratedAt string            `json:"generatedAt"`
	DB          database.Database `json:"db"`
}

type Catalog struct {
	mu       sync.RWMutex
	manifest Manifest
	loaded   bool
}

func NewCatalog() *Catalog {
	return &Catalog{}
}

func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	return manifest, nil
}

func WriteManifestAtomic(path string, manifest Manifest) error {
	data, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

func (c *Catalog) Set(manifest Manifest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.loaded && manifest.Version < c.manifest.Version {
		return
	}
	c.manifest = manifest
	c.loaded = true
}

func (c *Catalog) Manifest() (Manifest, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.manifest, c.loaded
}

func (c *Catalog) HasLevel(name string) bool {
	manifest, ok := c.Manifest()
	if !ok {
		return false
	}
	for _, item := range manifest.DB.Levels {
		if item.Name == name {
			return true
		}
	}
	return false
}
