package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/WindowsSov8forUs/sonolus-core-go/core"
	"github.com/WindowsSov8forUs/sonolus-core-go/database"
)

var levelNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type LevelUpload struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Artists     string   `json:"artists"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Rating      int      `json:"rating"`
	Engine      string   `json:"engine"`
	Cover       []byte   `json:"cover"`
	BGM         []byte   `json:"bgm"`
	Data        []byte   `json:"data"`
	Chart       []byte   `json:"chart,omitempty"`
}

var ErrLevelExists = fmt.Errorf("level already exists")

func WriteLevelSource(sourceDir string, upload LevelUpload) error {
	if upload.Name != "" && levelNamePattern.MatchString(upload.Name) {
		if _, err := os.Stat(filepath.Join(sourceDir, "levels", upload.Name)); err == nil {
			return ErrLevelExists
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err := validateLevelUpload(upload); err != nil {
		return err
	}

	levelDir := filepath.Join(sourceDir, "levels", upload.Name)
	if err := os.Mkdir(levelDir, 0o755); err != nil {
		if os.IsExist(err) {
			return ErrLevelExists
		}
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(levelDir), 0o755); err != nil {
			return err
		}
		if err := os.Mkdir(levelDir, 0o755); err != nil {
			if os.IsExist(err) {
				return ErrLevelExists
			}
			return err
		}
	}

	item := database.DatabaseLevelItem{
		Name:          upload.Name,
		Version:       1,
		Rating:        float64(upload.Rating),
		Title:         localized(upload.Title),
		Artists:       localized(upload.Artists),
		Author:        localized(upload.Author),
		Tags:          uploadTags(upload.Tags),
		Description:   localized(upload.Description),
		Engine:        upload.Engine,
		UseSkin:       database.DatabaseUseItem{UseDefault: true},
		UseBackground: database.DatabaseUseItem{UseDefault: true},
		UseEffect:     database.DatabaseUseItem{UseDefault: true},
		UseParticle:   database.DatabaseUseItem{UseDefault: true},
	}
	if err := writeJSON(filepath.Join(levelDir, "item.json"), item); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(levelDir, "cover.png"), upload.Cover, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(levelDir, "bgm.mp3"), upload.BGM, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(levelDir, "data.json"), upload.Data, 0o644); err != nil {
		return err
	}
	if len(upload.Chart) != 0 {
		if err := os.WriteFile(filepath.Join(levelDir, "chart.json"), upload.Chart, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func validateLevelUpload(upload LevelUpload) error {
	if upload.Name == "" || !levelNamePattern.MatchString(upload.Name) {
		return fmt.Errorf("name contains invalid characters")
	}
	if upload.Title == "" {
		return fmt.Errorf("title is required")
	}
	if upload.Artists == "" {
		return fmt.Errorf("artists is required")
	}
	if upload.Author == "" {
		return fmt.Errorf("author is required")
	}
	if upload.Engine == "" {
		return fmt.Errorf("engine is required")
	}
	if upload.Rating <= 0 {
		return fmt.Errorf("rating must be positive")
	}
	if len(upload.Cover) == 0 {
		return fmt.Errorf("cover is required")
	}
	if len(upload.BGM) == 0 {
		return fmt.Errorf("bgm is required")
	}
	if len(upload.Data) == 0 {
		return fmt.Errorf("data is required")
	}
	return nil
}

func localized(value string) database.LocalizationText {
	return database.LocalizationText{"en": core.Text(value), "zhs": core.Text(value), "zht": core.Text(value)}
}

func uploadTags(values []string) []database.DatabaseTag {
	result := make([]database.DatabaseTag, 0, len(values))
	for _, value := range values {
		result = append(result, database.DatabaseTag{Title: localized(value)})
	}
	return result
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
