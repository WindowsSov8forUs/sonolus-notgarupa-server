package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRebuildCreatesManifestAndRepository(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	createMinimalSource(t, source)

	manifest, err := Rebuild(context.Background(), RebuildOptions{
		SourceDir: source,
		DataDir:   filepath.Join(root, "data"),
		TmpDir:    filepath.Join(root, "tmp"),
		PublicURL: "https://repo.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Version == 0 || manifest.GeneratedAt == "" {
		t.Fatalf("manifest metadata missing: %#v", manifest)
	}
	if manifest.DB.Info.Banner == nil || manifest.DB.Info.Banner.URL == nil {
		t.Fatalf("banner missing: %#v", manifest.DB.Info.Banner)
	}
	url, _ := manifest.DB.Info.Banner.URL.Value()
	if url == "" || url[:len("https://repo.example/")] != "https://repo.example/" {
		t.Fatalf("banner url = %q", url)
	}
	if _, err := os.Stat(filepath.Join(root, "data", "manifest.json")); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(root, "data", "repository"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected repository blobs")
	}
}

func TestRebuildFailurePreservesOldManifest(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	createMinimalSource(t, source)
	dataDir := filepath.Join(root, "data")
	tmpDir := filepath.Join(root, "tmp")

	first, err := Rebuild(context.Background(), RebuildOptions{SourceDir: source, DataDir: dataDir, TmpDir: tmpDir, PublicURL: "https://repo.example"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "info.json"), []byte(`{"title":{"en":1}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Rebuild(context.Background(), RebuildOptions{SourceDir: source, DataDir: dataDir, TmpDir: tmpDir, PublicURL: "https://repo.example"}); err == nil {
		t.Fatal("expected rebuild failure")
	}
	var stored Manifest
	data, err := os.ReadFile(filepath.Join(dataDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatal(err)
	}
	if stored.Version != first.Version {
		t.Fatalf("manifest version = %d, want preserved %d", stored.Version, first.Version)
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
