package main

import (
	"embed"
	"opensplit-racetimegg/securestore"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// TODO:
	// Convert client_id and client_secret to live site (AFTER getting approval from racetime.gg staff)
	// Create an instance of the app structure
	app := NewApp("http", "localhost:8000", "localhost:9999")

	// TODO: switch to environment variable
	// app.encryptionKey = securestore.KeyFromEnv(os.Getenv("RACETIME_KEY"))
	app.encryptionKey = securestore.KeyFromEnv("TEST_KEY")

	// TODO: handle error
	app.Token, _ = securestore.LoadToken("token.enc", app.encryptionKey)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "opensplit-racetimegg",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
