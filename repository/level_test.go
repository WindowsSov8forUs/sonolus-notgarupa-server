package repository

import (
	"os"
	"path/filepath"
	"testing"
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
