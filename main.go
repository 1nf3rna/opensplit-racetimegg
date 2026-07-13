package main

import (
	"embed"
	"log"

	"opensplit-racetimegg/app"
	"opensplit-racetimegg/logger"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var mainLog = logger.Module("main")

func main() {
	logger.Init()

	client, err := app.New()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	runErr := wails.Run(&options.App{
		Title:     "opensplit-racetimegg",
		Width:     1024,
		Height:    768,
		MinWidth:  900,
		MinHeight: 580,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{
			R: 27,
			G: 38,
			B: 54,
			A: 1,
		},
		OnStartup: client.Startup,
		Bind: []any{
			client,
		},
	})

	if runErr != nil {
		mainLog.Error("wails.Run failed: %v", runErr)
	}
}
