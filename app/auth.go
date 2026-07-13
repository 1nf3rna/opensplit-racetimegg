package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"opensplit-racetimegg/racetime"
	"opensplit-racetimegg/securestore"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/oauth2"
)

func (a *App) Authorize() {
	url := a.conf.AuthCodeURL("state", oauth2.AccessTypeOnline, oauth2.S256ChallengeOption(a.verifier))

	codeChan := make(chan string)

	log.Info("AuthURL: %v", url)

	mux := http.NewServeMux()

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		go func() {
			codeChan <- code
		}()
	})

	server := &http.Server{
		Addr:    ":9999",
		Handler: mux,
	}
	defer server.Shutdown(a.ctx)

	go server.ListenAndServe()

	runtime.BrowserOpenURL(a.ctx, url)

	a.authCode = <-codeChan

	log.Info("AuthURL: %v", a.authCode)
}

// Requests tokens from authorization code
// Access tokens expire after 10 hours
// Can only be done if the user is authorized. Creates access and refresh tokens that needs to be stored. Expires eventually and needs to be refreshed with the refresh token.
// Example response should include: access_token, refresh_token, token_type, expires_in, scope
func (a *App) GenTokens() {
	ctx := context.Background()

	if len(a.authCode) == 0 {
		return
	}

	tok, err := a.conf.Exchange(ctx, a.authCode, oauth2.VerifierOption(a.verifier))

	if err != nil {
		log.Error("Generate Token Error: %v.\n", err)
	}

	a.Token = tok

	// TODO: Remove debug statements
	log.Info(
		"token refreshed access_expiry=%s refresh_present=%v",
		a.Token.Expiry.Format(time.RFC3339),
		a.Token.RefreshToken != "",
	)

	a.logToken()

	securestore.SaveToken("token.enc", *a.Token, a.encryptionKey)
}

// Can only be done if the user is logged in. Refreshes tokens that needs to be stored.
// Example response should include: access_token, refresh_token, token_type, expires_in, scope
func (a *App) refreshTokens() error {
	ctx := context.Background()

	ts := a.conf.TokenSource(ctx, a.Token)

	tok, err := ts.Token()
	if err != nil {
		log.Error("refresh failed:", err)

		errStr := strings.ToLower(err.Error())

		// Common OAuth invalid refresh token cases
		if strings.Contains(errStr, "invalid_grant") ||
			strings.Contains(errStr, "invalid_token") ||
			strings.Contains(errStr, "expired") ||
			strings.Contains(errStr, "revoked") {

			a.invalidateAuth("Session expired. Please log in again.")
		}

		return err
	}

	a.Token = tok

	a.logToken()

	err = securestore.SaveToken("token.enc", *a.Token, a.encryptionKey)
	if err != nil {
		log.Error("failed to save refreshed token:", err)
		return err
	}

	return nil
}

func (a *App) CheckTokens() string {
	log.Info("CheckTokens called")

	if a.Token == nil {
		log.Info("Token is nil")
		return ""
	}

	if a.Token.Valid() {
		log.Info("Token valid")
		return a.getAccessToken()
	}

	log.Info("Token expired, refreshing")

	// Access token expired

	// Refresh token missing
	if a.Token.RefreshToken == "" {
		a.invalidateAuth("No refresh token available.")
		return ""
	}

	// Refresh token exists, try using it
	err := a.refreshTokens()
	if err != nil {
		// invalid/revoked refresh token already handled
		return ""
	}

	// Refresh succeeded
	if a.Token != nil && a.Token.Valid() {
		return a.getAccessToken()
	}

	return ""
}

func (a *App) logToken() {
	if a.Token == nil {
		return
	}

	log.Debug("Access token: %s", a.Token.AccessToken)
	log.Debug("Refresh token: %s", a.Token.RefreshToken)
	log.Debug("Token type: %s", a.Token.TokenType)
	log.Debug("Expires: %s", a.Token.Expiry)
	log.Debug("Access token expires: %s", a.Token.ExpiresIn)
}

func (a *App) getAccessToken() (accessToken string) {
	go a.getUserInfo()
	return a.Token.AccessToken
}

func (a *App) invalidateAuth(reason string) {
	log.Info("authentication invalid:", reason)

	a.Token = nil

	// Remove encrypted token file
	err := securestore.DeleteToken("token.enc")
	if err != nil {
		log.Error("failed to delete token:", err)
	}

	// Notify frontend
	runtime.EventsEmit(a.ctx, "authExpired", reason)

	// Disconnect websocket if connected
	a.DisconnectRace()
}

func (a *App) getUserInfo() {
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(a.Token))

	resp, err := client.Get(racetime.WebRaceServer + "/o/userinfo")
	if err != nil {
		log.Error("userinfo request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Reading userinfo response body failed: %v", err)
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		a.invalidateAuth("Authentication expired.")
		return
	}

	var user racetime.UserInfo

	if err := json.Unmarshal(body, &user); err != nil {
		log.Error("Unmarshaling userinfo response body failed: %v", err)
	}

	log.Info("%+v\n", user)

	a.User = user
}
