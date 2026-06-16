package repository

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/WindowsSov8forUs/sonolus-core-go/core"
	"github.com/WindowsSov8forUs/sonolus-core-go/database"
	"github.com/WindowsSov8forUs/sonolus-pack-go/resource"
	"github.com/WindowsSov8forUs/sonolus-pack-go/schema"
)

func AppendPackedLevel(sourceDir string, packDir string, db database.Database, name string) (Snapshot, error) {
	for _, level := range db.Levels {
		if level.Name == name {
			return Snapshot{}, ErrLevelExists
		}
	}

	item, err := packLevel(sourceDir, packDir, name)
	if err != nil {
		return Snapshot{}, err
	}
	db.Levels = append(db.Levels, item)
	data, err := json.Marshal(db)
	if err != nil {
		return Snapshot{}, err
	}
	if err := WriteDatabaseAtomic(filepath.Join(packDir, "db.json"), data); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Version: time.Now().UnixMilli(), DB: db}, nil
}

func packLevel(sourceDir string, packDir string, name string) (database.DatabaseLevelItem, error) {
	dir := filepath.Join(sourceDir, "levels", name)
	item, err := schema.ParseLevelItem(filepath.Join(dir, "item.json"))
	if err != nil {
		return database.DatabaseLevelItem{}, err
	}
	item.Name = name

	packer := resource.Packer{Output: packDir, Logger: logger{}}
	if item.Cover, err = requiredSRL(packer.Pack(filepath.Join(dir, "cover"), "png", false)); err != nil {
		return database.DatabaseLevelItem{}, err
	}
	if item.BGM, err = requiredSRL(packer.Pack(filepath.Join(dir, "bgm"), "mp3", false)); err != nil {
		return database.DatabaseLevelItem{}, err
	}
	if srl, ok, err := packer.Pack(filepath.Join(dir, "preview"), "mp3", true); err != nil {
		return database.DatabaseLevelItem{}, err
	} else if ok {
		item.Preview = srl
	}
	if item.Data, err = requiredSRL(packer.Pack(filepath.Join(dir, "data"), "json", false)); err != nil {
		return database.DatabaseLevelItem{}, err
	}

	data, err := json.Marshal(item)
	if err != nil {
		return database.DatabaseLevelItem{}, err
	}
	var dbItem database.DatabaseLevelItem
	if err := json.Unmarshal(data, &dbItem); err != nil {
		return database.DatabaseLevelItem{}, err
	}
	return dbItem, nil
}

func requiredSRL(srl *core.Srl, ok bool, err error) (core.Srl, error) {
	if err != nil {
		return core.Srl{}, err
	}
	if !ok || srl == nil {
		return core.Srl{}, fmt.Errorf("required resource is missing")
	}
	return *srl, nil
}
