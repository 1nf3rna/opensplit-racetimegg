package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"opensplit-racetimegg/securestore"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/oauth2"
)

type App struct {
	Token         *oauth2.Token
	verifier      string
	conf          *oauth2.Config
	ctx           context.Context
	authMutex     sync.Mutex
	authCode      string
	encryptionKey []byte
	// codeChan  chan string
}

func NewApp(RestProtocol string, WebRaceServer string, RedirectURL string) *App {
	return &App{
		verifier: oauth2.GenerateVerifier(),
		conf: &oauth2.Config{
			ClientID:     "x4oiff8OAiWwtfQUboFhFlYfgmDMHmxduOFOQgve",
			ClientSecret: "1BYxBFqyO495W8VCYiZxAEXgortlLa5trpzY0xxDHNAuAWaqfxhgy4435Gq5yp6P76Hw1EIFdp8JjnKvDtDfzLZ2lo6D1TrrWlp0yNbmBTPpNxYVePSqE7eX72ZDAmaU",
			Scopes:       []string{"read", "chat_message", "race_action"},
			// RedirectURL:  RestProtocol + "://" + RedirectURL + "/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL:  RestProtocol + "://" + WebRaceServer + "/o/authorize",
				TokenURL: RestProtocol + "://" + WebRaceServer + "/o/token",
			},
		},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// // OAUTH_REDIRECT_ADDRESS
// 127.0.0.1

// // OAUTH_REDIRECT_PORT
// 4888

// // OAUTH_SCOPES
// read chat_message race_action

// // OAUTH_SERVER
// https://racetime.gg/

// // PROTOCOL_REST (http or https)
// https

// // PROTOCOL_WEBSOCKET (ws or wss)
// wss

// // domain (Domain or IP of the Race-Server)
// racetime.gg

func (a *App) Authorize() {
	url := a.conf.AuthCodeURL("state", oauth2.AccessTypeOnline, oauth2.S256ChallengeOption(a.verifier))

	codeChan := make(chan string)

	// fmt.Printf("URL for the auth dialog: %v\n", url)

	server := &http.Server{
		Addr: ":9999",
	}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		go func() {
			codeChan <- code
		}()
	})

	go server.ListenAndServe()

	runtime.BrowserOpenURL(a.ctx, url)

	a.authCode = <-codeChan

	server.Shutdown(a.ctx)
}

// Requests tokens from authorization code
// Access tokens expire after 10 hours
// Can only be done if the user is authorized. Creates access and refresh tokens that needs to be stored. Expires eventually and needs to be refreshed with the refresh token.
// Example response should include: access_token, refresh_token, token_type, expires_in, scope
// func (a *App) GenTokens() (accessToken string, refreshToken string) {
func (a *App) GenTokens() {
	ctx := context.Background()

	if len(a.authCode) == 0 {
		return
	}

	tok, err := a.conf.Exchange(ctx, a.authCode, oauth2.VerifierOption(a.verifier))

	if err != nil {
		log.Fatal(err)
	}

	a.Token = tok

	// TODO: Remove debug statements
	fmt.Printf("Access token: %s\n", a.Token.AccessToken)
	fmt.Printf("Refresh token: %s\n", a.Token.RefreshToken)
	fmt.Printf("Token type: %s\n", a.Token.TokenType)
	fmt.Printf("Access token expires: %s\n", a.Token.Expiry)
	fmt.Printf("Access token expires: %v\n", a.Token.ExpiresIn)

	securestore.SaveToken("token.enc", *a.Token, a.encryptionKey)
}

// Can only be done if the user is logged in. Refreshes tokens that needs to be stored.
// Example response should include: access_token, refresh_token, token_type, expires_in, scope
func (a *App) refreshTokens() {
	ctx := context.Background()

	// TODO: catch errors
	// no token, auth revoked
	a.conf.TokenSource(ctx, a.Token)

	// TODO: Remove debug statements
	fmt.Printf("Access token: %s\n", a.Token.AccessToken)
	fmt.Printf("Refresh token: %s\n", a.Token.RefreshToken)
	fmt.Printf("Token type: %s\n", a.Token.TokenType)
	fmt.Printf("Access token expires: %s\n", a.Token.Expiry)
	fmt.Printf("Access token expires: %v\n", a.Token.ExpiresIn)

	securestore.SaveToken("token.enc", *a.Token, a.encryptionKey)
}

func (a *App) CheckTokens() (accessToken string) {
	if a.Token == nil || (a.Token.RefreshToken == "" && a.Token.AccessToken == "") {
		return ""
	}
	if !a.Token.Valid() {
		if a.Token.RefreshToken != "" {
			a.refreshTokens()
			return a.getAccessToken()
		} else {
			return ""
		}
	}
	return a.getAccessToken()
}

func (a *App) getAccessToken() (accessToken string) {
	return a.Token.AccessToken
}
