package serverinfo

import (
	"github.com/WindowsSov8forUs/sonolus-core-go/core"
	coreserver "github.com/WindowsSov8forUs/sonolus-core-go/core/server"
	sonolus "github.com/WindowsSov8forUs/sonolus-server-go"
)

func Install(app *sonolus.Sonolus) {
	app.ServerInfoHandler = func(ctx sonolus.Context) (sonolus.ServerInfoModel, *sonolus.ServerError) {
		var buttons []coreserver.ServerInfoButton
		if len(app.Post.Items) > 0 {
			buttons = append(buttons, coreserver.ServerInfoItemButton{Type: core.ItemTypePost})
		}
		if len(app.Playlist.Items) > 0 {
			buttons = append(buttons, coreserver.ServerInfoItemButton{Type: core.ItemTypePlaylist})
		}
		buttons = append(buttons, coreserver.ServerInfoItemButton{Type: core.ItemTypeLevel})
		if len(app.Replay.Items) > 0 {
			buttons = append(buttons, coreserver.ServerInfoItemButton{Type: core.ItemTypeReplay})
		}
		if len(app.Skin.Items) > 0 {
			buttons = append(buttons, coreserver.ServerInfoItemButton{Type: core.ItemTypeSkin})
		}
		if len(app.Background.Items) > 0 {
			buttons = append(buttons, coreserver.ServerInfoItemButton{Type: core.ItemTypeBackground})
		}
		if len(app.Effect.Items) > 0 {
			buttons = append(buttons, coreserver.ServerInfoItemButton{Type: core.ItemTypeEffect})
		}
		if len(app.Particle.Items) > 0 {
			buttons = append(buttons, coreserver.ServerInfoItemButton{Type: core.ItemTypeParticle})
		}
		if len(app.Engine.Items) > 0 {
			buttons = append(buttons, coreserver.ServerInfoItemButton{Type: core.ItemTypeEngine})
		}
		buttons = append(buttons, coreserver.ServerInfoConfigurationButton{Type: "configuration"})

		model := sonolus.ServerInfoModel{
			Title:       app.Title,
			Description: app.Description,
			Buttons:     buttons,
			Banner:      app.Banner,
		}
		model.Configuration.Options = app.Configuration.Options
		return model, nil
	}
}
