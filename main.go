package main

import (
	"embed"
	"opensplit-racetimegg/logger"
	"opensplit-racetimegg/securestore"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var mainLog = logger.Module("main")

func main() {
	logger.Init()

	// Create an instance of the app structure
	app, err := NewApp()
	if err != nil {
		log.Fatal("failed to initialize app: %v", err)
	}

	// TODO: switch to environment variable
	// app.encryptionKey = securestore.KeyFromEnv(os.Getenv("RACETIME_KEY"))
	app.encryptionKey = securestore.KeyFromEnv("TEST_KEY")

	// TODO: handle error
	app.Token, _ = securestore.LoadToken("token.enc", app.encryptionKey)

	// Create application with options
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
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if runErr != nil {
		mainLog.Error("wails.Run failed: %v", runErr)
	}
}
