package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"opensplit-racetimegg/securestore"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/oauth2"
)

const socketUrl = "ws://localhost:9999"

type RaceState int

const (
	invitational = iota
	pending
	partitioned //(only for ladder 1v1 races)
	open
	in_progress
	finished
	cancelled
)

type UserRole int

const (
	Unknown UserRole = iota
	Anonymous
	Regular
	ChannelCreator UserRole = 4
	Monitor        UserRole = 8
	Moderator      UserRole = 16
	Staff          UserRole = 32
	Bot            UserRole = 64
	System         UserRole = 128
)

type App struct {
	Token                *oauth2.Token
	verifier             string
	conf                 *oauth2.Config
	ctx                  context.Context
	authMutex            sync.Mutex
	authCode             string
	encryptionKey        []byte
	racetimeWS           *websocket.Conn
	authenticatedRaceURL string
	handlers             map[string]func([]byte)
	writeCh              chan []byte
}

func NewApp(RestProtocol string, WebRaceServer string, RedirectURL string) *App {
	client := &App{
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
		writeCh:  make(chan []byte, 10000000),
		handlers: map[string]func([]byte){},
	}

	// Register handlers
	client.handlers["chat.message"] = client.HandleChatMessage
	client.handlers["chat.history"] = client.HandleChatHistory
	client.handlers["chat.dm"] = client.HandleChatDM
	client.handlers["chat.pin"] = client.HandleChatPin
	client.handlers["chat.unpin"] = client.HandleChatUnpin
	client.handlers["chat.delete"] = client.HandleChatDelete
	client.handlers["chat.purge"] = client.HandleChatPurge
	client.handlers["error"] = client.HandleChatError
	client.handlers["pong"] = client.HandlePong
	client.handlers["race.data"] = client.HandleRaceData

	return client
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

type messageData struct {
	//	{
	Action string `json:"type"` //	  "action": "message",
	//	  "data": {
	Message   string  `json:"data.message"`             //	    "message": "Your message goes here",
	Pinned    bool    `json:"data.pinned"`              //	    "pinned": <bool>,
	Actions   *string `json:"data.actions,omitempty"`   //	    "actions": <object or null>,
	Direct_to *string `json:"data.direct_to,omitempty"` //	    "direct_to": <string or null>,
	Guid      string  `json:"data.guid"`                //	    "guid": "<random string>"
	//	  }
	//	}
}

func (a *App) SendText(text string) {
	a.Send(messageData{
		Action: "message",
	})
}

type BaseMessage struct {
	Type string `json:"type"`
}

type ChatMessage struct {
	//	{
	Type string `json:"type"` //	  "type": "chat.message",
	//	  "message": {
	ID string `json:"message.id"` //	    "id": "<string>",
	//	    "user": { <user info object> },
	//	    "bot": "<string>",
	//	    "direct_to": { <user info object> },
	//	    "posted_at": "<iso date string>"
	//	    "message": "<string>",
	//	    "message_plain": "<string>",
	//	    "highlight": <bool>,
	//	    "is_dm": <bool>,
	//	    "is_bot": <bool>,
	//	    "is_system": <bool>,
	//	    "is_pinned": <bool>,
	//	    "delay": "<iso duration string>",
	//	    "actions" { <action objects> }
	//	  }
	//	}
}

func (a *App) HandleChatMessage(data []byte) {
	var msg ChatMessage

	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Println("chat decode error:", err)
		return
	}

	fmt.Printf(
		"[CHAT] %s: %s\n",
		msg.ID,
		// msg.Message,
	)
	//     const hours: number = event.data.message.posted_at.getHours();
	// // Get the minutes (0-59)
	// const minutes: number = event.data.message.posted_at.getMinutes();
	// // You can then format them as needed, for example, with leading zeros
	// const formattedHours: string = String(hours).padStart(2, '0');
	// const formattedMinutes: string = String(minutes).padStart(2, '0');

	// console.log(`Hours: ${hours}`);
	// console.log(`Minutes: ${minutes}`);
	// console.log(`Formatted Time: ${formattedHours}:${formattedMinutes}`);

	// if (paragraph) {
	//     paragraph.textContent += formattedHours + ":" + formattedMinutes + " " + event.data.message.user.name + event.data.message.message
	// }
}

type ChatHistory struct {
	//	{
	//	  "type": "chat.history",
	//	  "messages": [
	//	     {"id":"xa2wrRW32bl48fJq", ...},
	//	     {"id":"g6Kem5bewJfG3ds2", ...},
	//	  ]
	//	}
}

func (a *App) HandleChatHistory(data []byte) {
	//   if (paragraph) {
	//             for (let index = 0; index < event.data.messages.length; index++) {
	//                 paragraph.textContent += event.data.messages[index]
	//             }
	//         }
}

type ChatDM struct {
	//	{
	//	  "type": "chat.dm",
	//	  "message": "<string>",
	//	  "from_user": { <user info object> },
	//	  "from_bot": "<string>",
	//	  "to": { <user info object> },
	//	}
}

func (a *App) HandleChatDM(data []byte) {
}

type ChatPin struct {
	// {
	//   "type": "chat.pin",
	//   "message": { ... }
	// }
}

func (a *App) HandleChatPin(data []byte) {
}

type ChatUnpin struct {
	// {
	//   "type": "chat.unpin",
	//   "message": { ... }
	// }
}

func (a *App) HandleChatUnpin(data []byte) {
}

type ChatDelete struct {
	// chat.delete
	// {
	//     "type": "chat.delete",
	//     "delete": {
	//         "id": "<string>",
	//         "user": { <user info object> },
	//         "bot": "<string>",
	//         "is_bot": <bool>,
	//         "deleted_by": { <user info object> }
	//     }
	// }
}

func (a *App) HandleChatDelete(data []byte) {
}

type ChatPurge struct {
	// {
	// "type": "chat.purge",
	// "purge": {
	// "user": { <user info object> },
	// "purged_by": { <user info object> }
	// }
	// }
}

func (a *App) HandleChatPurge(data []byte) {
}

type ChatError struct {
	// {
	//   "type": "error",
	//   "errors": [
	// "Permission denied, you may need to re-authorise this application.",
	// "..."
	//   ]
	// }
}

func (a *App) HandleChatError(data []byte) {
}

type Pong struct {
	// {
	//   "type": "pong"
	// }
}

func (a *App) HandlePong(data []byte) {
}

type RaceData struct {
	// {
	//   "type": "race.data",
	//   "race": {
	// ...
	//   }
	// }
}

// This will be received:
// 1) On first connect
// 2) After a system message
// 3) After a getrace action
func (a *App) HandleRaceData(data []byte) {
	//         // goal = event.data.race.goal.name
	//         // info = event.data.race.info
	//         // entrants = event.data.race.entrants
	//         // category = event.data.race.category.name
	//         // raceID = event.data.race.slug
}

// true for forfeit; false for unforfeit
func (a *App) Forfeit(state bool) {
	// // message format
	// // {
	// //     "action": "message",
	// //     "data": {
	// //         "message": "Your message goes here",
	// //         // "pinned": <bool>,
	// //         "actions": <object or null>,
	// //         "direct_to": <string or null>,
	// //         "guid": "<random string>"
	// //     }
	// // }
	// if (ws.readyState === WebSocket.OPEN) {
	//     console.log('Forfeit status changed!');
	//     // if forfeited unforfeit otherwise forfeit
	//     if (forfeit) {
	//         const mData: messageData = {
	//             message: ".unforfeit"
	//         }
	//         const ready_message: { action: string; data: messageData } = {
	//             action: "message",
	//             data: mData
	//         }

	//         forfeit = !forfeit
	//         ws.send(JSON.stringify(ready_message));
	//     } else {
	//         const mData: messageData = {
	//             message: ".forfeit"
	//         }
	//         const ready_message: { action: string; data: messageData } = {
	//             action: "message",
	//             data: mData
	//         }

	//         forfeit = !forfeit
	//         ws.send(JSON.stringify(ready_message));
}

// true for done; false for undone
func (a *App) Done(state bool) {
	// // message format
	// // {
	// //     "action": "message",
	// //     "data": {
	// //         "message": "Your message goes here",
	// //         // "pinned": <bool>,
	// //         "actions": <object or null>,
	// //         "direct_to": <string or null>,
	// //         "guid": "<random string>"
	// //     }
	// // }
	// if (ws.readyState === WebSocket.OPEN) {
	//     console.log('Race join status changed!');
	//     // if done undone otherwise done
	//     if (done) {
	//         const mData: messageData = {
	//             message: ".undone"
	//         }
	//         const ready_message: { action: string; data: messageData } = {
	//             action: "message",
	//             data: mData
	//         }

	//         done = !done
	//         ws.send(JSON.stringify(ready_message));
	//     } else {
	//         const mData: messageData = {
	//             message: ".done"
	//         }
	//         const ready_message: { action: string; data: messageData } = {
	//             action: "message",
	//             data: mData
	//         }

	//         done = !done
	//         ws.send(JSON.stringify(ready_message));
}

// true for ready; false for unready
func (a *App) Ready(state bool) {
	// 	// message format
	//     // {
	//     //     "action": "message",
	//     //     "data": {
	//     //         "message": "Your message goes here",
	//     //         // "pinned": <bool>,
	//     //         "actions": <object or null>,
	//     //         "direct_to": <string or null>,
	//     //         "guid": "<random string>"
	//     //     }
	//     // }

	//     const mData: messageData = {
	//         message: ".ready"
	//     }
	//     const ready_message: { action: string; data: messageData } = {
	//         action: "message",
	//         data: mData
	//     }

	//     ws.send(JSON.stringify(ready_message));
	// } else {
	//     console.log('Checkbox is unchecked')
	//     // message format
	//     // {
	//     //     "action": "message",
	//     //     "data": {
	//     //         "message": "Your message goes here",
	//     //         // "pinned": <bool>,
	//     //         "actions": <object or null>,
	//     //         "direct_to": <string or null>,
	//     //         "guid": "<random string>"
	//     //     }
	//     // }
	//     const mData: messageData = {
	//         message: ".unready"
	//     }
	//     const ready_message: { action: string; data: messageData } = {
	//         action: "message",
	//         data: mData
	//     }

	//     ws.send(JSON.stringify(ready_message));
}

// true for join; false for leave
func (a *App) Join(state bool) {
	// // message format
	// // {
	// //     "action": "message",
	// //     "data": {
	// //         "message": "Your message goes here",
	// //         // "pinned": <bool>,
	// //         "actions": <object or null>,
	// //         "direct_to": <string or null>,
	// //         "guid": "<random string>"
	// //     }
	// // }
	// if (ws.readyState === WebSocket.OPEN) {
	//     console.log('Race join status changed!');
	//     // if in race leave otherwise enter
	//     if (joined) {
	//         const mData: messageData = {
	//             message: ".leave"
	//         }
	//         const ready_message: { action: string; data: messageData } = {
	//             action: "message",
	//             data: mData
	//         }

	//         joined = !joined
	//         ws.send(JSON.stringify(ready_message));
	//     } else {
	//         const mData: messageData = {
	//             message: ".join"
	//         }
	//         const ready_message: { action: string; data: messageData } = {
	//             action: "message",
	//             data: mData
	//         }

	//         joined = !joined
	//         ws.send(JSON.stringify(ready_message));
}

// true for hide results; false for show results
func (a *App) UpdateEntrantList(state bool) {
	// List of entrants with stream status (color coded icon??), ready status (color code name??)
	//     const entrantList: HTMLUListElement = document.createElement('ul');
	// for (let index = 0; index < entrants.length; index++) {
	//     const element = entrants[index].name;

	//     const listItem: HTMLLIElement = document.createElement('li');
	//     listItem.textContent = element;
	//     entrantList.appendChild(listItem);
	// }

	// w.document.body.appendChild(entrantList);
}

func (a *App) SaveLog() {
	// save chat box text to file
}

func (a *App) ForceReload() {
	// no idea
}

// open websocket connection and start goroutines
func (a *App) WebSocketConnection(raceURL string) {
	// should probably do this with authorization header as shown here:
	// https://github.com/racetimeGG/racetime-app/wiki/Category-bots

	// TODO: Fix raceURL
	authenticatedRaceURL := "/ws/o/race/" + raceURL
	a.authenticatedRaceURL = socketUrl + authenticatedRaceURL + "?token=" + a.Token.AccessToken

	a.racetimeWS, _, _ = websocket.Dial(
		a.ctx,
		a.authenticatedRaceURL,
		nil,
	)

	go a.pingRoutine()
	go a.writeRoutine()
	go a.readRoutine()
}

// Convert data to be sent to json before sending
func (a *App) Send(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	select {
	case a.writeCh <- data:
		return nil

	case <-a.ctx.Done():
		return fmt.Errorf("connection closed")
	}
}

// message writing routine
func (a *App) writeRoutine() {
	for {
		select {
		case msg := <-a.writeCh:
			writeCtx, cancel := context.WithTimeout(
				a.ctx,
				5*time.Second,
			)

			err := a.racetimeWS.Write(
				writeCtx,
				websocket.MessageText,
				msg,
			)

			cancel()

			if err != nil {
				log.Println("write error:", err)
				// c.cancel()
				return
			}

		case <-a.ctx.Done():
			return
		}
	}
}

// message recieve and routing routine
func (a *App) readRoutine() {
	// defer c.cancel()

	for {
		_, data, err := a.racetimeWS.Read(a.ctx)
		if err != nil {
			log.Println("read error:", err)
			return
		}

		var base BaseMessage

		err = json.Unmarshal(data, &base)
		if err != nil {
			log.Println("invalid json:", err)
			continue
		}

		handler, ok := a.handlers[base.Type]
		if !ok {
			log.Println("unknown message type:", base.Type)
			continue
		}

		handler(data)
	}
}

// keep alive goroutine
func (a *App) pingRoutine() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Timeout for ping response
			pingCtx, cancel := context.WithTimeout(a.ctx, 5*time.Second)

			err := a.racetimeWS.Ping(pingCtx)

			cancel()

			if err != nil {
				log.Println("ping failed:", err)
				return
			}

			log.Println("ping sent")
		case <-a.ctx.Done():
			return
		}
	}
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
