package main

import (
	"context"
	"embed"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wangle201210/wachat/backend/config"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load configuration
	ctx := context.Background()
	cfg, err := config.Load(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "Failed to load configuration: %v", err)
	}

	// Create an instance of the app structure
	app := NewApp(cfg)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "wachat",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		g.Log().Fatal(ctx, err)
	}
}
