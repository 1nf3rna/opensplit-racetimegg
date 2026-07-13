package app

import (
	"context"
	"opensplit-racetimegg/logger"
	"opensplit-racetimegg/processing"
	"opensplit-racetimegg/racetime"
	"opensplit-racetimegg/securestore"

	"golang.org/x/oauth2"
)

var log = logger.Module("app").SetLevel(logger.InfoLevel)

// // OAUTH_REDIRECT_ADDRESS
// 127.0.0.1

// // OAUTH_REDIRECT_PORT
// 4888

type App struct {
	Token          *oauth2.Token
	verifier       string
	conf           *oauth2.Config
	ctx            context.Context
	authCode       string
	encryptionKey  []byte
	handlers       map[string]func([]byte)
	raceConn       *RaceConnection
	CurrentRace    racetime.RaceInfo
	User           racetime.UserInfo
	engine         *processing.Engine
	osConnectionCh chan bool
	canJoin        bool
}

func New() (*App, error) {
	engine, connCh, err := processing.NewEngine()
	if err != nil {
		return nil, err
	}

	app := &App{
		verifier: oauth2.GenerateVerifier(),
		conf: &oauth2.Config{
			ClientID:     "ILLY5XtgStv8Z3hsQNg8nvWg2f16y4Uau38MwBgD",
			ClientSecret: "pLvZfhr7NvcpQwJ5IkVvzgYqGh8WqVWcvmtiMCmy15jFhINotEcfjlkUb6L0WM7tYkt4aNjooyFGRsPo8GjBG0rDcewk40sMWbgnlbL67VnTCKWoVupGd5eJbx2gbQbW",
			Scopes:       []string{"read", "chat_message", "race_action"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  racetime.WebRaceServer + "/o/authorize",
				TokenURL: racetime.WebRaceServer + "/o/token",
			},
		},
		handlers:       map[string]func([]byte){},
		engine:         engine,
		osConnectionCh: connCh,
	}

	// TODO: switch to environment variable
	// app.encryptionKey = securestore.KeyFromEnv(os.Getenv("RACETIME_KEY"))
	app.encryptionKey = securestore.KeyFromEnv("TEST_KEY")

	// TODO: handle error
	app.Token, _ = securestore.LoadToken("token.enc", app.encryptionKey)

	app.registerHandlers()

	return app, nil
}

func (a *App) registerHandlers() {
	a.handlers = map[string]func([]byte){
		"chat.message": a.HandleChatMessage,
		"chat.history": a.HandleChatHistory,
		"chat.dm":      a.HandleChatDM,
		"chat.pin":     a.HandleChatPin,
		"chat.unpin":   a.HandleChatUnpin,
		"chat.delete":  a.HandleChatDelete,
		"chat.purge":   a.HandleChatPurge,
		"error":        a.HandleChatError,
		"pong":         a.HandlePong,
		"race.data":    a.HandleRaceData,
		"race.renders": a.HandleRenders,
	}
}
