package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/WindowsSov8forUs/sonolus-core-go/database"
)

func TestRebuildCreatesDatabaseAndRepository(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	createMinimalSource(t, source)

	snapshot, err := Rebuild(context.Background(), RebuildOptions{
		SourceDir: source,
		PackDir:   filepath.Join(root, "pack"),
		TmpDir:    filepath.Join(root, "tmp"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Version == 0 {
		t.Fatalf("snapshot version missing: %#v", snapshot)
	}
	if snapshot.DB.Info.Banner == nil || snapshot.DB.Info.Banner.URL == nil {
		t.Fatalf("banner missing: %#v", snapshot.DB.Info.Banner)
	}
	url, _ := snapshot.DB.Info.Banner.URL.Value()
	if url == "" || !strings.HasPrefix(url, "/sonolus/repository/") {
		t.Fatalf("banner url = %q, want repository path", url)
	}
	if _, err := os.Stat(filepath.Join(root, "pack", "db.json")); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(root, "pack", "repository"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected repository blobs")
	}
}

func TestRebuildFailurePreservesOldDatabase(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	createMinimalSource(t, source)
	packDir := filepath.Join(root, "pack")
	tmpDir := filepath.Join(root, "tmp")

	first, err := Rebuild(context.Background(), RebuildOptions{SourceDir: source, PackDir: packDir, TmpDir: tmpDir})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "info.json"), []byte(`{"title":{"en":1}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Rebuild(context.Background(), RebuildOptions{SourceDir: source, PackDir: packDir, TmpDir: tmpDir}); err == nil {
		t.Fatal("expected rebuild failure")
	}
	var stored database.Database
	data, err := os.ReadFile(filepath.Join(packDir, "db.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatal(err)
	}
	if stored.Info.Title["en"] != first.DB.Info.Title["en"] {
		t.Fatalf("db title = %#v, want preserved %#v", stored.Info.Title, first.DB.Info.Title)
	}
}

func TestCreatePackWorkDirUsesTimestamp(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 6, 17, 12, 34, 56, 789, time.UTC)
	dir, err := createPackWorkDir(root, now)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(dir) != "pack-20260617T123456.000000789Z" {
		t.Fatalf("work dir = %q, want timestamped pack dir", filepath.Base(dir))
	}
	again, err := createPackWorkDir(root, now)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(again) != "pack-20260617T123456.000000789Z-1" {
		t.Fatalf("collision dir = %q, want suffixed timestamped pack dir", filepath.Base(again))
	}
}

func createMinimalSource(t *testing.T, source string) {
	t.Helper()
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(source, "info.json"), `{"title":{"en":"Server"},"description":{"en":"Description"}}`)
	write(t, filepath.Join(source, "banner.png"), "banner")
}

func write(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
