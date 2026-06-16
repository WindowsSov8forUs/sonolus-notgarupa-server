package app

import (
	"context"
	"log"

	"github.com/WindowsSov8forUs/sonolus-core-go/core"
	"github.com/WindowsSov8forUs/sonolus-core-go/database"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/chart"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/config"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/level"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/middleware"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/repository"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/serverinfo"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/upload"
	sonolus "github.com/WindowsSov8forUs/sonolus-server-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func BuildRouter(cfg config.Config) (*gin.Engine, error) {
	app := sonolus.New(sonolus.Options{
		Address:        cfg.Address,
		FallbackLocale: "zhs",
		Upload: sonolus.UploadOptions{
			MaxMemory: 30 << 20,
			MaxCount:  1,
		},
	})
	seedBuiltinCatalog(app)

	store := repository.NewStore(repository.StoreConfig{
		SourceDir: cfg.RepositorySourceDir,
		DataDir:   cfg.RepositoryDataDir,
		TmpDir:    cfg.RepositoryTmpDir,
		PublicURL: cfg.RepositoryPublicURL,
	})
	catalog := repository.NewCatalog()
	registeredRepositoryFiles := map[string]bool{}
	if snapshot, err := store.Rebuild(context.Background()); err != nil {
		log.Printf("repository rebuild unavailable: %v", err)
		if snapshot, err := store.Snapshot(); err != nil {
			log.Printf("repository manifest unavailable: %v", err)
		} else {
			catalog.Set(snapshot.Manifest)
			applyRepository(app, snapshot, registeredRepositoryFiles)
		}
	} else {
		catalog.Set(snapshot.Manifest)
		applyRepository(app, snapshot, registeredRepositoryFiles)
	}
	if cfg.RepositoryWatchSource {
		store.StartWatcher(context.Background(), cfg.RepositoryPollInterval, func(snapshot repository.Snapshot) {
			catalog.Set(snapshot.Manifest)
			applyRepository(app, snapshot, registeredRepositoryFiles)
		})
	}

	level.NewServiceFromItems(func() []sonolus.LevelItemModel { return app.Level.Items }).Install(app)
	serverinfo.Install(app)

	router := gin.Default()
	router.Use(cors.Default())
	router.Use(middleware.RequestSizeLimiter(30 << 20))
	router.Use(middleware.RemoveSonolusVersionHeader())
	(&upload.Handler{
		Engines:   engineNames(app),
		Catalog:   catalog,
		Publisher: store,
		RefreshCatalog: func() {
			snapshot, err := store.Snapshot()
			if err != nil {
				log.Printf("refresh repository manifest after upload: %v", err)
				return
			}
			catalog.Set(snapshot.Manifest)
			applyRepository(app, snapshot, registeredRepositoryFiles)
		},
	}).Install(router)
	app.Install(router)

	return router, nil
}

func applyRepository(app *sonolus.Sonolus, snapshot repository.Snapshot, registered map[string]bool) {
	applyManifest(app, snapshot.Manifest)
	registerRepositoryFiles(app, snapshot.Blobs, registered)
}

func registerRepositoryFiles(app *sonolus.Sonolus, blobs []repository.BlobFile, registered map[string]bool) {
	for _, blob := range blobs {
		if registered[blob.Hash] {
			continue
		}
		if _, err := app.AddFile(blob.Path, blob.Hash); err != nil {
			log.Printf("register repository file %s: %v", blob.Hash, err)
			continue
		}
		registered[blob.Hash] = true
	}
}

func applyManifest(app *sonolus.Sonolus, manifest repository.Manifest) {
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

func engineNames(app *sonolus.Sonolus) map[string]bool {
	names := make(map[string]bool, len(app.Engine.Items))
	for _, engine := range app.Engine.Items {
		names[engine.Name] = true
	}
	return names
}

func seedBuiltinCatalog(app *sonolus.Sonolus) {
	app.Skin.Items = []sonolus.SkinItemModel{{
		DatabaseSkinItem: database.DatabaseSkinItem{
			Name:      "builtin-skin",
			Version:   1,
			Title:     localized("Builtin Skin"),
			Subtitle:  localized("Builtin"),
			Author:    localized("sonolus-notgarupa-server"),
			Thumbnail: srl("builtin-skin-thumbnail"),
			Data:      srl("builtin-skin-data"),
			Texture:   srl("builtin-skin-texture"),
		},
	}}
	app.Background.Items = []sonolus.BackgroundItemModel{{
		DatabaseBackgroundItem: database.DatabaseBackgroundItem{
			Name:          "builtin-background",
			Version:       1,
			Title:         localized("Builtin Background"),
			Subtitle:      localized("Builtin"),
			Author:        localized("sonolus-notgarupa-server"),
			Thumbnail:     srl("builtin-background-thumbnail"),
			Data:          srl("builtin-background-data"),
			Image:         srl("builtin-background-image"),
			Configuration: srl("builtin-background-configuration"),
		},
	}}
	app.Effect.Items = []sonolus.EffectItemModel{{
		DatabaseEffectItem: database.DatabaseEffectItem{
			Name:      "builtin-effect",
			Version:   1,
			Title:     localized("Builtin Effect"),
			Subtitle:  localized("Builtin"),
			Author:    localized("sonolus-notgarupa-server"),
			Thumbnail: srl("builtin-effect-thumbnail"),
			Data:      srl("builtin-effect-data"),
			Audio:     srl("builtin-effect-audio"),
		},
	}}
	app.Particle.Items = []sonolus.ParticleItemModel{{
		DatabaseParticleItem: database.DatabaseParticleItem{
			Name:      "builtin-particle",
			Version:   1,
			Title:     localized("Builtin Particle"),
			Subtitle:  localized("Builtin"),
			Author:    localized("sonolus-notgarupa-server"),
			Thumbnail: srl("builtin-particle-thumbnail"),
			Data:      srl("builtin-particle-data"),
			Texture:   srl("builtin-particle-texture"),
		},
	}}
	app.Engine.Items = []sonolus.EngineItemModel{
		builtinEngine(chart.EngineNotGarupa, "NotGarupa"),
		builtinEngine(chart.EngineNotGarupaHabahiro, "NotGarupa Habahiro"),
	}
}

func builtinEngine(name string, title string) sonolus.EngineItemModel {
	return sonolus.EngineItemModel{
		DatabaseEngineItem: database.DatabaseEngineItem{
			Name:          name,
			Version:       13,
			Title:         localized(title),
			Subtitle:      localized("Builtin"),
			Author:        localized("sonolus-notgarupa-server"),
			Skin:          "builtin-skin",
			Background:    "builtin-background",
			Effect:        "builtin-effect",
			Particle:      "builtin-particle",
			Thumbnail:     srl(name + "-thumbnail"),
			PlayData:      srl(name + "-play-data"),
			WatchData:     srl(name + "-watch-data"),
			PreviewData:   srl(name + "-preview-data"),
			TutorialData:  srl(name + "-tutorial-data"),
			Configuration: srl(name + "-configuration"),
		},
	}
}

func localized(value string) database.LocalizationText {
	return database.LocalizationText{"en": core.Text(value), "zhs": core.Text(value), "zht": core.Text(value)}
}

func srl(hash string) core.Srl {
	hashValue := core.Value(hash)
	urlValue := core.Value("/sonolus/repository/" + hash)
	return core.Srl{Hash: &hashValue, URL: &urlValue}
}
