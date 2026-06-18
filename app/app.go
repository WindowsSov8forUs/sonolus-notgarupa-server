package app

import (
	"context"
	"log"
	"net/http"

	"github.com/WindowsSov8forUs/sonolus-core-go/database"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/buildinfo"
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
		FallbackLocale: "zhs",
		Upload: sonolus.UploadOptions{
			MaxMemory: 30 << 20,
			MaxCount:  1,
		},
	})

	store := repository.NewStore(repository.StoreConfig{
		SourceDir: cfg.Repository.SourceDir,
		PackDir:   cfg.Repository.PackDir,
		TmpDir:    cfg.Repository.TmpDir,
	})
	levelNames := repository.NewLevelNameIndex()
	registeredRepositoryFiles := map[string]bool{}
	if snapshot, err := store.Rebuild(context.Background()); err != nil {
		log.Printf("repository rebuild unavailable: %v", err)
		if snapshot, err := store.Snapshot(); err != nil {
			log.Printf("repository database unavailable: %v", err)
		} else {
			levelNames.Set(snapshot)
			applyRepository(app, snapshot, registeredRepositoryFiles)
		}
	} else {
		levelNames.Set(snapshot)
		applyRepository(app, snapshot, registeredRepositoryFiles)
	}

	level.NewServiceFromItems(func() []sonolus.LevelItemModel { return app.Level.Items }).Install(app)
	serverinfo.Install(app)

	router := gin.Default()
	router.Use(cors.Default())
	router.Use(middleware.RequestSizeLimiter(30 << 20))
	router.Use(middleware.RemoveSonolusVersionHeader())
	router.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "sonolus-notgarupa-server\nVersion: %s\nStatus: OK\nSonolus API: /sonolus/info\n", buildinfo.Version)
	})
	router.HEAD("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	(&upload.Handler{
		Engines:    engineNames(app),
		LevelNames: levelNames,
		Publisher:  store,
		RefreshSnapshot: func(snapshot repository.Snapshot) {
			levelNames.Set(snapshot)
			applyRepository(app, snapshot, registeredRepositoryFiles)
		},
	}).Install(router)
	app.Install(router)

	return router, nil
}

func applyRepository(app *sonolus.Sonolus, snapshot repository.Snapshot, registered map[string]bool) {
	applyDatabase(app, snapshot.DB)
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

func applyDatabase(app *sonolus.Sonolus, db database.Database) {
	app.Title = db.Info.Title
	app.Description = db.Info.Description
	app.Banner = db.Info.Banner
	app.Post.Items = mapItems(db.Posts, func(item database.DatabasePostItem) sonolus.PostItemModel {
		return sonolus.PostItemModel{DatabasePostItem: item}
	})
	app.Playlist.Items = mapItems(db.Playlists, func(item database.DatabasePlaylistItem) sonolus.PlaylistItemModel {
		return sonolus.PlaylistItemModel{DatabasePlaylistItem: item}
	})
	app.Level.Items = mapItems(db.Levels, func(item database.DatabaseLevelItem) sonolus.LevelItemModel {
		return sonolus.LevelItemModel{DatabaseLevelItem: item}
	})
	app.Skin.Items = mapItems(db.Skins, func(item database.DatabaseSkinItem) sonolus.SkinItemModel {
		return sonolus.SkinItemModel{DatabaseSkinItem: item}
	})
	app.Background.Items = mapItems(db.Backgrounds, func(item database.DatabaseBackgroundItem) sonolus.BackgroundItemModel {
		return sonolus.BackgroundItemModel{DatabaseBackgroundItem: item}
	})
	app.Effect.Items = mapItems(db.Effects, func(item database.DatabaseEffectItem) sonolus.EffectItemModel {
		return sonolus.EffectItemModel{DatabaseEffectItem: item}
	})
	app.Particle.Items = mapItems(db.Particles, func(item database.DatabaseParticleItem) sonolus.ParticleItemModel {
		return sonolus.ParticleItemModel{DatabaseParticleItem: item}
	})
	app.Engine.Items = mapItems(db.Engines, func(item database.DatabaseEngineItem) sonolus.EngineItemModel {
		return sonolus.EngineItemModel{DatabaseEngineItem: item}
	})
	app.Replay.Items = mapItems(db.Replays, func(item database.DatabaseReplayItem) sonolus.ReplayItemModel {
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
