package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/WindowsSov8forUs/sonolus-core-go/database"
)

type LevelNameIndex struct {
	mu      sync.RWMutex
	version int64
	names   map[string]struct{}
	loaded  bool
}

func NewLevelNameIndex() *LevelNameIndex {
	return &LevelNameIndex{}
}

func LoadDatabase(path string) (database.Database, int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return database.Database{}, 0, err
	}
	var db database.Database
	if err := json.Unmarshal(data, &db); err != nil {
		return database.Database{}, 0, fmt.Errorf("decode db: %w", err)
	}
	version := int64(0)
	if info, err := os.Stat(path); err == nil {
		version = info.ModTime().UnixMilli()
	}
	return db, version, nil
}

func WriteDatabaseAtomic(path string, data []byte) error {
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

func (i *LevelNameIndex) Set(snapshot Snapshot) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.loaded && snapshot.Version < i.version {
		return
	}
	names := make(map[string]struct{}, len(snapshot.DB.Levels))
	for _, item := range snapshot.DB.Levels {
		names[item.Name] = struct{}{}
	}
	i.version = snapshot.Version
	i.names = names
	i.loaded = true
}

func (i *LevelNameIndex) HasLevel(name string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if !i.loaded {
		return false
	}
	_, ok := i.names[name]
	return ok
}
