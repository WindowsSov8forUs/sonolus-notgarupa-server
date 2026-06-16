package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/WindowsSov8forUs/sonolus-core-go/database"
)

func TestWriteLevelSource(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source")
	err := WriteLevelSource(source, LevelUpload{
		Name:        "level-1",
		Title:       "Title",
		Artists:     "Artists",
		Author:      "Author",
		Description: "Description",
		Tags:        []string{"tag"},
		Rating:      20,
		Engine:      "engine",
		Cover:       []byte("cover"),
		BGM:         []byte("bgm"),
		Data:        []byte(`{"entities":[]}`),
		Chart:       []byte(`[]`),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"item.json", "cover.png", "bgm.mp3", "data.json", "chart.json"} {
		if _, err := os.Stat(filepath.Join(source, "levels", "level-1", name)); err != nil {
			t.Fatalf("%s missing: %v", name, err)
		}
	}
	if err := WriteLevelSource(source, LevelUpload{Name: "level-1"}); err != ErrLevelExists {
		t.Fatalf("collision err = %v, want ErrLevelExists", err)
	}
}

func TestUploadLevelAppendsPackedLevelToDatabase(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	createMinimalPlayableSource(t, source)
	store := NewStore(StoreConfig{
		SourceDir: source,
		PackDir:   filepath.Join(root, "pack"),
		TmpDir:    filepath.Join(root, "tmp"),
	})
	first, err := store.Rebuild(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(first.DB.Levels) != 0 {
		t.Fatalf("initial levels=%d, want 0", len(first.DB.Levels))
	}

	snapshot, err := store.UploadLevel(context.Background(), LevelUpload{
		Name:        "level-1",
		Title:       "Title",
		Artists:     "Artists",
		Author:      "Author",
		Description: "Description",
		Tags:        []string{"tag"},
		Rating:      20,
		Engine:      "engine",
		Cover:       []byte("cover"),
		BGM:         []byte("bgm"),
		Data:        []byte(`{"entities":[]}`),
		Chart:       []byte(`[]`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.DB.Levels) != 1 || snapshot.DB.Levels[0].Name != "level-1" {
		t.Fatalf("levels=%#v, want uploaded level", snapshot.DB.Levels)
	}
	if len(snapshot.Blobs) == 0 {
		t.Fatal("expected repository blobs")
	}

	data, err := os.ReadFile(filepath.Join(root, "pack", "db.json"))
	if err != nil {
		t.Fatal(err)
	}
	var stored database.Database
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatal(err)
	}
	if len(stored.Levels) != 1 || stored.Levels[0].Name != "level-1" {
		t.Fatalf("stored levels=%#v, want uploaded level", stored.Levels)
	}
}

func createMinimalPlayableSource(t *testing.T, source string) {
	t.Helper()
	write(t, filepath.Join(source, "info.json"), `{"title":{"en":"Server"}}`)
	writeItemFiles(t, filepath.Join(source, "skins", "skin"), `{"version":4,"title":{"en":"Skin"},"subtitle":{"en":"Skin"},"author":{"en":"Author"},"tags":[]}`, map[string]string{
		"thumbnail": "skin-thumbnail",
		"data":      `{}`,
		"texture":   "skin-texture",
	})
	writeItemFiles(t, filepath.Join(source, "backgrounds", "background"), `{"version":2,"title":{"en":"Background"},"subtitle":{"en":"Background"},"author":{"en":"Author"},"tags":[]}`, map[string]string{
		"thumbnail":     "background-thumbnail",
		"data":          `{}`,
		"image":         "background-image",
		"configuration": `{}`,
	})
	writeItemFiles(t, filepath.Join(source, "effects", "effect"), `{"version":5,"title":{"en":"Effect"},"subtitle":{"en":"Effect"},"author":{"en":"Author"},"tags":[]}`, map[string]string{
		"thumbnail": "effect-thumbnail",
		"data":      `{}`,
		"audio":     "effect-audio",
	})
	writeItemFiles(t, filepath.Join(source, "particles", "particle"), `{"version":3,"title":{"en":"Particle"},"subtitle":{"en":"Particle"},"author":{"en":"Author"},"tags":[]}`, map[string]string{
		"thumbnail": "particle-thumbnail",
		"data":      `{}`,
		"texture":   "particle-texture",
	})
	writeItemFiles(t, filepath.Join(source, "engines", "engine"), `{"version":13,"title":{"en":"Engine"},"subtitle":{"en":"Engine"},"author":{"en":"Author"},"tags":[],"skin":"skin","background":"background","effect":"effect","particle":"particle"}`, map[string]string{
		"thumbnail":     "engine-thumbnail",
		"playData":      `{}`,
		"watchData":     `{}`,
		"previewData":   `{}`,
		"tutorialData":  `{}`,
		"configuration": `{}`,
	})
}

func writeItemFiles(t *testing.T, dir string, item string, files map[string]string) {
	t.Helper()
	write(t, filepath.Join(dir, "item.json"), item)
	for name, content := range files {
		write(t, filepath.Join(dir, name), content)
	}
}
