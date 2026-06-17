package repository

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/WindowsSov8forUs/sonolus-core-go/database"
)

type Store struct {
	cfg StoreConfig

	mu sync.Mutex
}

type StoreConfig struct {
	SourceDir string
	PackDir   string
	TmpDir    string
}

type BlobFile struct {
	Hash string
	Path string
}

type Snapshot struct {
	Version int64
	DB      database.Database
	Blobs   []BlobFile
}

func NewStore(cfg StoreConfig) *Store {
	return &Store{cfg: cfg}
}

func (s *Store) Rebuild(ctx context.Context) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot, err := Rebuild(ctx, RebuildOptions{
		SourceDir: s.cfg.SourceDir,
		PackDir:   s.cfg.PackDir,
		TmpDir:    s.cfg.TmpDir,
		Logger:    logger{},
	})
	if err != nil {
		return Snapshot{}, err
	}
	return s.withBlobs(snapshot)
}

func (s *Store) UploadLevel(ctx context.Context, upload LevelUpload) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := WriteLevelSource(s.cfg.SourceDir, upload); err != nil {
		return Snapshot{}, err
	}
	db, _, err := LoadDatabase(filepath.Join(s.cfg.PackDir, "db.json"))
	if err != nil {
		snapshot, rebuildErr := Rebuild(ctx, RebuildOptions{
			SourceDir: s.cfg.SourceDir,
			PackDir:   s.cfg.PackDir,
			TmpDir:    s.cfg.TmpDir,
			Logger:    logger{},
		})
		if rebuildErr != nil {
			return Snapshot{}, err
		}
		return s.withBlobs(snapshot)
	}
	snapshot, err := AppendPackedLevel(s.cfg.SourceDir, s.cfg.PackDir, db, upload.Name)
	if err != nil {
		return Snapshot{}, err
	}
	return s.withBlobs(snapshot)
}

func (s *Store) Snapshot() (Snapshot, error) {
	db, version, err := LoadDatabase(filepath.Join(s.cfg.PackDir, "db.json"))
	if err != nil {
		return Snapshot{}, err
	}
	return s.withBlobs(Snapshot{Version: version, DB: db})
}

func (s *Store) withBlobs(snapshot Snapshot) (Snapshot, error) {
	blobs, err := blobFiles(filepath.Join(s.cfg.PackDir, "repository"))
	if err != nil {
		return Snapshot{}, err
	}
	snapshot.Blobs = blobs
	return snapshot, nil
}

func blobFiles(dir string) ([]BlobFile, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []BlobFile{}, nil
	}
	if err != nil {
		return nil, err
	}
	blobs := make([]BlobFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		hash := entry.Name()
		blobs = append(blobs, BlobFile{
			Hash: hash,
			Path: filepath.Join(dir, hash),
		})
	}
	return blobs, nil
}

type logger struct{}

func (logger) Info(args ...any)    { log.Print(args...) }
func (logger) Warning(args ...any) { log.Print(args...) }
