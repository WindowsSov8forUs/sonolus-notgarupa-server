package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/WindowsSov8forUs/sonolus-core-go/database"
	sonolus "github.com/WindowsSov8forUs/sonolus-server-go"
)

type Manifest struct {
	Version     int64             `json:"version"`
	GeneratedAt string            `json:"generatedAt"`
	DB          database.Database `json:"db"`
}

type Catalog struct {
	manifestURL string
	client      *http.Client
	mu          sync.RWMutex
	manifest    Manifest
	loaded      bool
}

func NewCatalog(manifestURL string) *Catalog {
	return &Catalog{
		manifestURL: manifestURL,
		client:      &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Catalog) Refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.manifestURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch repository manifest: status=%d", resp.StatusCode)
	}
	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return err
	}
	c.Set(manifest)
	return nil
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

func (c *Catalog) StartPolling(interval time.Duration, apply func(Manifest)) {
	if interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			err := c.Refresh(ctx)
			cancel()
			if err != nil {
				continue
			}
			if manifest, ok := c.Manifest(); ok {
				apply(manifest)
			}
		}
	}()
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

func ApplyManifest(app *sonolus.Sonolus, manifest Manifest) {
	app.Title = manifest.DB.Info.Title
	app.Description = manifest.DB.Info.Description
	app.Banner = manifest.DB.Info.Banner
	app.Post.Items = mapItems(manifest.DB.Posts, func(item database.DatabasePostItem) sonolus.PostItemModel {
		return sonolus.PostItemModel{DatabasePostItem: item}
	})
	app.Playlist.Items = mapItems(manifest.DB.Playlists, func(item database.DatabasePlaylistItem) sonolus.PlaylistItemModel {
		return sonolus.PlaylistItemModel{DatabasePlaylistItem: item}
	})
	app.Level.Items = mapItems(manifest.DB.Levels, func(item database.DatabaseLevelItem) sonolus.LevelItemModel {
		return sonolus.LevelItemModel{DatabaseLevelItem: item}
	})
	app.Skin.Items = mapItems(manifest.DB.Skins, func(item database.DatabaseSkinItem) sonolus.SkinItemModel {
		return sonolus.SkinItemModel{DatabaseSkinItem: item}
	})
	app.Background.Items = mapItems(manifest.DB.Backgrounds, func(item database.DatabaseBackgroundItem) sonolus.BackgroundItemModel {
		return sonolus.BackgroundItemModel{DatabaseBackgroundItem: item}
	})
	app.Effect.Items = mapItems(manifest.DB.Effects, func(item database.DatabaseEffectItem) sonolus.EffectItemModel {
		return sonolus.EffectItemModel{DatabaseEffectItem: item}
	})
	app.Particle.Items = mapItems(manifest.DB.Particles, func(item database.DatabaseParticleItem) sonolus.ParticleItemModel {
		return sonolus.ParticleItemModel{DatabaseParticleItem: item}
	})
	app.Engine.Items = mapItems(manifest.DB.Engines, func(item database.DatabaseEngineItem) sonolus.EngineItemModel {
		return sonolus.EngineItemModel{DatabaseEngineItem: item}
	})
	app.Replay.Items = mapItems(manifest.DB.Replays, func(item database.DatabaseReplayItem) sonolus.ReplayItemModel {
		return sonolus.ReplayItemModel{DatabaseReplayItem: item}
	})
}

func mapItems[S any, T any](items []S, fn func(S) T) []T {
	result := make([]T, 0, len(items))
	for _, item := range items {
		result = append(result, fn(item))
	}
	return result
}
