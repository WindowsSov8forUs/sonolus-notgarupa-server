package app

import (
	"context"
	"log"
	"time"

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

	var catalog *repository.Catalog
	if cfg.RepositoryManifestURL != "" {
		catalog = repository.NewCatalog(cfg.RepositoryManifestURL)
		if err := catalog.Refresh(context.Background()); err != nil {
			log.Printf("repository manifest unavailable: %v", err)
		} else if manifest, ok := catalog.Manifest(); ok {
			repository.ApplyManifest(app, manifest)
		}
		catalog.StartPolling(cfg.RepositoryPollInterval, func(manifest repository.Manifest) {
			repository.ApplyManifest(app, manifest)
		})
	}

	level.NewServiceFromItems(func() []sonolus.LevelItemModel { return app.Level.Items }).Install(app)
	serverinfo.Install(app)

	router := gin.Default()
	router.Use(cors.Default())
	router.Use(middleware.RequestSizeLimiter(30 << 20))
	router.Use(middleware.RemoveSonolusVersionHeader())
	(&upload.Handler{
		Engines:    engineNames(app),
		Catalog:    catalog,
		Repository: repositoryClient(cfg.RepositoryAdminURL),
		RefreshCatalog: func() {
			if catalog == nil {
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if err := catalog.Refresh(ctx); err != nil {
				log.Printf("refresh repository manifest after upload: %v", err)
				return
			}
			if manifest, ok := catalog.Manifest(); ok {
				repository.ApplyManifest(app, manifest)
			}
		},
	}).Install(router)
	app.Install(router)

	return router, nil
}

func repositoryClient(adminURL string) *repository.Client {
	if adminURL == "" {
		return nil
	}
	return repository.NewClient(adminURL)
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
