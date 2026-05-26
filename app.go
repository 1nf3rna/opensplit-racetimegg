package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"opensplit-racetimegg/processing"
	"opensplit-racetimegg/securestore"
	"os"

	"strings"
	"sync"
	"time"

	duration "github.com/channelmeter/iso8601duration"
	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/oauth2"
)

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

// local dev
const socketUrl = "ws://localhost:8000"
const WebRaceServer = "http://localhost:8000"

// live
// const socketUrl = "wss://racetime.gg"
// const WebRaceServer = "https://racetime.gg"

// type RaceState int

// const (
// 	invitational = iota
// 	pending
// 	partitioned //(only for ladder 1v1 races)
// 	open
// 	inProgressState
// 	finished
// 	cancelled
// )

// type UserRole int

// const (
// 	Unknown UserRole = iota
// 	Anonymous
// 	Regular
// 	ChannelCreator UserRole = 4
// 	Monitor        UserRole = 8
// 	Moderator      UserRole = 16
// 	Staff          UserRole = 32
// 	Bot            UserRole = 64
// 	System         UserRole = 128
// )

// type UserStatus int

// const (
// 	ready = iota
// 	not_ready
// 	inProgressStatus
// 	done
// 	dnf //(did not finish, i.e. forfeited)
// 	dq  //(disqualified)
// 	requested
// 	invited
// 	declined
// )

type UserInfo struct {
	ID            string `json:"id"`
	FullName      string `json:"full_name"`
	Name          string `json:"name"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	TwitchName    string `json:"twitch_name"`
	IsStaff       bool   `json:"is_staff"`
}

type RaceInfo struct {
	Version              int
	Goal                 string
	Game                 string
	RaceID               string
	Info                 string
	DisplayResults       bool
	EntrantCount         int
	EntrantFinishedCount int
	EntrantInactiveCount int
	Entrants             []Entrant
	Text                 []ChatMessage
	Ranked               bool
	AutoStart            bool
	Delay                int64
	Status               string
	StatusVerbose        string
	StatusHelpText       string
	DisqualifyUnready    bool
}

type Entrant struct {
	User           User          `json:"user"` // user: User data blob for this entrant.
	Status         EntrantStatus `json:"status"`
	Place          *int          `json:"place"`           // place: Integer indicating what position the user finished in.
	PlaceOrdinal   *string       `json:"place_ordinal"`   // place_ordinal: String ordinal version of place, e.g. "3rd".
	Score          *int          `json:"score"`           // score: Integer amount of points earned by this entrant on the relevant leaderboard. Note that this is not the entrant's current score (unless the race is in progress), it is the score they had when they entered the race, not after.
	ScoreChange    *int          `json:"score_change"`    // score_change Integer amount of points gained/lost as a result of this race, or null (not zero!) if race is not recorded.
	Comment        *string       `json:"comment"`         // comment: A string containing a pithy comeback supplied by the user post-race, or null if they have no comment. If hide_comments is true and the race has not concluded, this field is always null.
	HasComment     *bool         `json:"has_comment"`     // has_comment: A boolean indicating if the entrant has made a comment. This field is unaffected by the hide_comments setting.
	StreamLive     bool          `json:"stream_live"`     // stream_live: Boolean indicating if the user's stream is currently live. This is updated in real-time while a race is in progress, but once an entrant has finished, forfeited or been disqualified it will not be updated.
	StreamOverride bool          `json:"stream_override"` // stream_override: Boolean indicating if a moderator overrode the streaming requirement for this race entrant,
}

type EntrantStatus struct {
	Value        string `json:"value"`         // value: A machine-parsable status text.
	VerboseValue string `json:"verbose_value"` // verbose_value: A user-parsable status text, e.g. "In progress".
	HelpText     string `json:"help_text"`     // help_text: Describes the status, e.g. "Did not finish the race.".
	// ISO8601 duration string
	FinishTime *string    `json:"finish_time"` // finish_time: The user's final finish time, or null if they've not finished (ISO 8601 duration).
	FinishedAt *time.Time `json:"finished_at"` // finished_at: The date/time when the user finished, or null if they've not finished (ISO 8601 date).
}

type Category struct {
	Name      string `json:"name"`       // name: The name of the category, e.g. "Super Mario 64".
	ShortName string `json:"short_name"` // short_name: An abbreviated name, e.g. "OoTR".
	Slug      string `json:"slug"`       // slug: Unique category slug (part of the URL).
	URL       string `json:"url"`        // url: URL for the main category page.
	DataURL   string `json:"data_url"`   // data_url: URL for the category data endpoint, which you can use to obtain more detailed category information.
}

type RaceStatus struct {
	State        string `json:"value"`         // value: A machine-parsable status text. Possible values are:
	VerboseValue string `json:"verbose_value"` // verbose_value: A user-parsable status text, e.g. "In progress".
	HelpText     string `json:"help_text"`     // help_text: Describes the status, e.g. "Race is in progress".
}

type Goal struct {
	Name   string `json:"name"`   // name: A string value indicating the current goal.
	Custom bool   `json:"custom"` // custom: A boolean indicating if the goal name was custom, or one of the pre-set category goals.
}

type User struct {
	Id             string `json:"id"`             // "id": "fR42gLweew3pQlm4",
	Full_name      string `json:"full_name"`      // "full_name": "Mario#5527",
	Name           string `json:"name"`           // "name": "Mario",
	Discriminator  string `json:"discriminator"`  // "discriminator": "5527",
	Url            string `json:"url"`            // "url": "/user/fR42gLweew3pQlm4",
	Avatar         string `json:"avatar"`         // "avatar": "/media/mario.png",
	Pronouns       string `json:"pronouns"`       // "pronouns": "he/him",
	Flair          string `json:"flair"`          // "flair": "monitor supporter",
	Twitch_name    string `json:"twitch_name"`    // "twitch_name": "ItsaMeMario",
	Twitch_channel string `json:"twitch_channel"` // "twitch_channel": "https://www.twitch.tv/itsamemario",
	Can_moderate   bool   `json:"can_moderate"`   // "can_moderate": false
}

type ConnectionStatus byte

const (
	Disconnected   ConnectionStatus = 0
	Connected      ConnectionStatus = 1
	Reconnecting   ConnectionStatus = 2
	WaitingForGame ConnectionStatus = 3
)

type ConnectionState struct {
	ConnectionStatus ConnectionStatus `json:"connection_status"`
	Message          string           `json:"message"`
}

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
	CurrentRace          RaceInfo
	User                 UserInfo
	engine               *processing.Engine
	osConnectionCh       chan bool
	logger               *Logger
}

type Logger struct {
	debug bool
}

func NewLogger(debug bool) *Logger {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	return &Logger{
		debug: debug,
	}
}

func (l *Logger) Debug(format string, v ...any) {
	if !l.debug {
		return
	}

	log.Printf("[DEBUG] "+format, v...)
}

func (l *Logger) Info(format string, v ...any) {
	log.Printf("[INFO] "+format, v...)
}

func (l *Logger) Warn(format string, v ...any) {
	log.Printf("[WARN] "+format, v...)
}

func (l *Logger) Error(format string, v ...any) {
	log.Printf("[ERROR] "+format, v...)
}

func NewApp() *App {
	engine, connCh := processing.NewEngine()

	client := &App{
		verifier: oauth2.GenerateVerifier(),
		conf: &oauth2.Config{
			// local dev
			ClientID:     "x4oiff8OAiWwtfQUboFhFlYfgmDMHmxduOFOQgve",
			ClientSecret: "1BYxBFqyO495W8VCYiZxAEXgortlLa5trpzY0xxDHNAuAWaqfxhgy4435Gq5yp6P76Hw1EIFdp8JjnKvDtDfzLZ2lo6D1TrrWlp0yNbmBTPpNxYVePSqE7eX72ZDAmaU",
			// Live racetime.gg
			// ClientID:     "ILLY5XtgStv8Z3hsQNg8nvWg2f16y4Uau38MwBgD",
			// ClientSecret: "pLvZfhr7NvcpQwJ5IkVvzgYqGh8WqVWcvmtiMCmy15jFhINotEcfjlkUb6L0WM7tYkt4aNjooyFGRsPo8GjBG0rDcewk40sMWbgnlbL67VnTCKWoVupGd5eJbx2gbQbW",
			Scopes: []string{"read", "chat_message", "race_action"},
			// RedirectURL:  RestProtocol + "://" + RedirectURL + "/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL:  WebRaceServer + "/o/authorize",
				TokenURL: WebRaceServer + "/o/token",
			},
		},
		writeCh:        make(chan []byte, 2000),
		handlers:       map[string]func([]byte){},
		engine:         engine,
		osConnectionCh: connCh,
		logger:         NewLogger(os.Getenv("DEBUG") == "1"),
	}
	client.CurrentRace.DisplayResults = true

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
	client.handlers["race.renders"] = client.HandleRenders

	return client
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	go func() {
		s := ConnectionState{}
		for {
			status, ok := <-a.osConnectionCh
			if !ok {
				return
			}

			s.ConnectionStatus = Disconnected
			s.Message = "OpenSplit Not Found"
			if status {
				s.ConnectionStatus = Connected
				s.Message = "OpenSplit Connected"
			}

			runtime.EventsEmit(a.ctx, "opensplit:connection", s)
		}
	}()

	go func() {
		for {
			ev, ok := <-a.engine.Events()
			if !ok {
				return
			}

			switch ev.Command {
			case processing.DONE:
				a.logger.Info("opensplit DONE event received")
				a.SendText(".done", a.generateGUID())

			case processing.UNDONE:
				a.logger.Info("opensplit UNDONE event received")
				a.SendText(".undone", a.generateGUID())
			}
		}
	}()
}

func (a *App) generateGUID() string {
	return uuid.NewString()
}

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

type MessageDataEnvelope struct {
	Action string      `json:"action"`
	Data   MessageData `json:"data"`
}

type MessageData struct {
	Message   string  `json:"message,omitempty"`   //	    "message": "Your message goes here",
	Pinned    bool    `json:"pinned,omitempty"`    //	    "pinned": <bool>,
	Actions   *string `json:"actions,omitempty"`   //	    "actions": <object or null>,
	Direct_to *string `json:"direct_to,omitempty"` //	    "direct_to": <string or null>,
	GUID      string  `json:"guid"`                //	    "guid": "<random string>"
}

func (a *App) SendText(text string, GUID string) {
	a.logger.Debug("SendText called")

	a.Send(MessageDataEnvelope{
		Action: "message",
		Data: MessageData{
			Message: text,
			Pinned:  false,
			GUID:    GUID,
		},
	})
}

type BaseMessage struct {
	Type string `json:"type"`
}

type DurationString string

func (d *DurationString) UnmarshalJSON(data []byte) error {
	// null
	if string(data) == "null" {
		*d = ""
		return nil
	}

	// string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*d = DurationString(s)
		return nil
	}

	// number
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*d = DurationString(fmt.Sprintf("%d", n))
		return nil
	}

	return fmt.Errorf("invalid delay format: %s", string(data))
}

type ChatMessageEnvelope struct {
	// Type    string      `json:"type"`
	Message ChatMessage `json:"message"`
}

type ChatMessage struct {
	ID            string         `json:"id"`            // "id": "<string>",
	User          User           `json:"user"`          // "user": { <user info object> },
	Bot           *string        `json:"bot"`           // "bot": "<string>",
	DirectTo      *User          `json:"direct_to"`     // "direct_to": { <user info object> },
	PostedAt      time.Time      `json:"posted_at"`     // "posted_at": "<iso date string>"
	Message       string         `json:"message"`       // "message": "<string>",
	Message_plain string         `json:"message_plain"` // "message_plain": "<string>",
	Highlight     bool           `json:"highlight"`     // "highlight": <bool>,
	Is_dm         bool           `json:"is_dm"`         // "is_dm": <bool>,
	Is_bot        bool           `json:"is_bot"`        // "is_bot": <bool>,
	Is_system     bool           `json:"is_system"`     // "is_system": <bool>,
	Is_pinned     bool           `json:"is_pinned"`     // "is_pinned": <bool>,
	Delay         DurationString `json:"delay"`         // "delay": "<iso duration string>",
	//	    "actions" { <action objects> }
}

func (a *App) HandleChatMessage(data []byte) {
	var env ChatMessageEnvelope

	err := json.Unmarshal(data, &env)
	if err != nil {
		a.logger.Error("chat decode error: %v", err)
		return
	}

	msg := env.Message

	fmt.Printf("ChatMessage\n")
	fmt.Printf("[CHAT] %+v\n", msg)

	// ignore duplicate messages
	for _, m := range a.CurrentRace.Text {
		if m.ID == msg.ID {
			return
		}
	}

	a.CurrentRace.Text = append(a.CurrentRace.Text, msg)

	// Notify frontend
	runtime.EventsEmit(a.ctx, "chatUpdated", a.CurrentRace.Text)
}

type ChatHistory struct {
	// Type string `json:"type"` // "type": "chat.history",
	Messages []ChatMessage `json:"messages"`
	//	  "messages": [
	//	     {"id":"xa2wrRW32bl48fJq", ...},
	//	     {"id":"g6Kem5bewJfG3ds2", ...},
	//	  ]
}

func (a *App) HandleChatHistory(data []byte) {
	var msg ChatHistory

	err := json.Unmarshal(data, &msg)
	if err != nil {
		a.logger.Error("chat decode error: %v", err)

		return
	}

	fmt.Printf("ChatHistory\n")

	fmt.Printf("[CHAT] %+v\n", msg)

	// replace race message array
	a.CurrentRace.Text = msg.Messages

	// Notify frontend
	runtime.EventsEmit(a.ctx, "chatUpdated", a.CurrentRace.Text)
}

type ChatDM struct {
	// Type     string `json:"type"`      //	  "type": "chat.dm",
	Message  string `json:"message"`   //	  "message": "<string>",
	FromUser User   `json:"from_user"` //	  "from_user": { <user info object> },
	From_bot string `json:"from_bot"`  //	  "from_bot": "<string>",
	To       User   `json:"to"`        //	  "to": { <user info object> },
}

func (a *App) HandleChatDM(data []byte) {
	var msg ChatDM

	err := json.Unmarshal(data, &msg)
	if err != nil {
		a.logger.Error("chat decode error: %v", err)

		return
	}

	fmt.Printf("ChatDM\n")

	fmt.Printf("[CHAT] %+v\n", msg)

	// This message type doesn't matter
}

type ChatPin struct {
	// Type string `json:"type"` //   "type": "chat.pin",
	//   "message": { ... }
	Message ChatMessage `json:"message"`
}

func (a *App) HandleChatPin(data []byte) {
	var msg ChatPin

	err := json.Unmarshal(data, &msg)
	if err != nil {
		a.logger.Error("chat decode error: %v", err)

		return
	}

	fmt.Printf("ChatPin\n")
	fmt.Printf("[CHAT] %+v\n", msg)

	// handle pinning message to top of chat window
	for i, m := range a.CurrentRace.Text {
		if m.ID == msg.Message.ID {
			a.CurrentRace.Text[i].Is_pinned = true

			// Notify frontend
			runtime.EventsEmit(a.ctx, "chatUpdated", a.CurrentRace.Text)

			return
		}
	}

	runtime.EventsEmit(a.ctx, "chatUpdated", a.CurrentRace.Text)
}

type ChatUnpin struct {
	// Type string `json:"type"` //   "type": "chat.unpin",
	//   "message": { ... }
	Message ChatMessage `json:"message"`
}

func (a *App) HandleChatUnpin(data []byte) {
	var msg ChatUnpin

	err := json.Unmarshal(data, &msg)
	if err != nil {
		a.logger.Error("chat decode error: %v", err)

		return
	}

	fmt.Printf("ChatUnpin\n")
	fmt.Printf("[CHAT] %+v\n", msg)

	// handle unpinning message from the top of chat window
	for i, m := range a.CurrentRace.Text {
		if m.ID == msg.Message.ID {
			a.CurrentRace.Text[i].Is_pinned = false

			// Notify frontend
			runtime.EventsEmit(a.ctx, "chatUpdated", a.CurrentRace.Text)

			return
		}
	}

	runtime.EventsEmit(a.ctx, "chatUpdated", a.CurrentRace.Text)
}

type ChatDeleteEnvelope struct {
	// Type   string     `json:"type"`
	Delete ChatDelete `json:"delete"`
}

type ChatDelete struct {
	ID        string  `json:"id"`         //         "id": "<string>",
	User      *User   `json:"user"`       //         "user": { <user info object> },
	Bot       *string `json:"bot"`        //         "bot": "<string>",
	Is_bot    bool    `json:"is_bot"`     //         "is_bot": <bool>,
	DeletedBy User    `json:"deleted_by"` //         "deleted_by": { <user info object> }
}

func (a *App) HandleChatDelete(data []byte) {
	var env ChatDeleteEnvelope

	err := json.Unmarshal(data, &env)
	if err != nil {
		a.logger.Error("chat decode error: %v", err)

		return
	}

	msg := env.Delete

	fmt.Printf("ChatDelete\n")
	fmt.Printf("[CHAT] %+v\n", msg)

	for i, m := range a.CurrentRace.Text {
		if m.ID == msg.ID {
			// Remove element at index i
			a.CurrentRace.Text = append(a.CurrentRace.Text[:i], a.CurrentRace.Text[i+1:]...)

			// Notify frontend
			runtime.EventsEmit(a.ctx, "chatUpdated", a.CurrentRace.Text)

			return
		}
	}
}

type ChatPurgeEnvelope struct {
	// Type  string    `json:"type"`
	Purge ChatPurge `json:"purge"`
}

type ChatPurge struct {
	User     User `json:"user"`      // "user": { <user info object> },
	PurgedBy User `json:"purged_by"` // "purged_by": { <user info object> }
}

func (a *App) HandleChatPurge(data []byte) {
	var env ChatPurgeEnvelope

	err := json.Unmarshal(data, &env)
	if err != nil {
		a.logger.Error("chat decode error: %v", err)

		return
	}

	msg := env.Purge

	fmt.Printf("ChatPurge\n")
	fmt.Printf("[CHAT] %+v\n", msg)

	filtered := a.CurrentRace.Text[:0]

	for _, m := range a.CurrentRace.Text {
		if m.User.Id != msg.User.Id {
			filtered = append(filtered, m)
		}
	}

	a.CurrentRace.Text = filtered

	// Notify frontend
	runtime.EventsEmit(a.ctx, "chatUpdated", a.CurrentRace.Text)
}

type ChatError struct {
	// {
	// Type string `json:"type"` //   "type": "error",
	Errors []string `json:"errors"`
	//   "errors": [
	// "Permission denied, you may need to re-authorise this application.",
	// "..."
	//   ]
	// }
}

func (a *App) HandleChatError(data []byte) {
	var msg ChatError

	err := json.Unmarshal(data, &msg)
	if err != nil {
		a.logger.Error("chat decode error: %v", err)

		return
	}

	fmt.Printf("ChatError\n")

	fmt.Printf("[CHAT] %+v\n", msg)

	// Do stuff depending on the errors
	for _, msgError := range msg.Errors {
		switch msgError {
		case "You are not eligible to join this race.":
			// Streaming required and twitch channel not linked (join   request_to_join   invite)
			fmt.Println("User not eligible to join race")

			// Disable join button
			runtime.EventsEmit(a.ctx, "joinEligibility", false)
		case "Races cannot have more than 5 monitors.":
			// Inform user too many monitors
		case "Ensure this value has at most 1000 characters (it has 52428).":
			// Inform user message is too long
		case "Race is not an invitational.":
			// Set if race is being changed from invitational to open when not in that state
		case "Race is not open.":
			// Set if race is being changed from open to invitational when not in that state
		case "Race cannot be started yet.":
			// Set trying to start race while conditions don't allow it to start (can_begin)
		case "Cannot cancel a race that is in %(state)s state.":
			// Set when trying to cancel a done race
		case "Race cannot be partitioned yet.":
			// Set when trying to partition a race (can_partition)
		case "Cannot finish a race that has not been started.":
			// Set when trying to finish a race that's not in progress (is_in_progress  finish)
		case "Cannot restart a race from this state.":
			//(is_unfinalized  unfinish)
		case "Race cannot be finalized, it is on hold.":
			//(hold   (un)record)
		case "This race cannot be recorded because one or more entrants have deleted their account. Please set this race to \"Do not record\".":
		case "Race is not recordable or already recorded.":
			//((un)record)
		case "Race hold cannot be changed now.":
			// (add/remove hold)
		case "Unable to comply, racing in progress.":
			// Set when race in progress and room opener; can't make rematch
		case "Only race monitors may create a rematch. Start a new race room instead.":
			// Set when trying to rematch when not a race monitor
		case "You are not allowed to start a new race.":
			// User not allowed to make races
		case "Not a team race.":
			// Not a team race (create team   join team   get_available_teams)
		case "Cannot join a team (join the race first!).":
			// Join race first
		case "You are already in that team.":
			// Cannot join team multiple times
		case "Cannot change team during the race.":
		case "You cannot join that team without an invitation.":
		case "You are not allowed to quit this race.":
			// invite or joined and disqualify_unready enabled (decline_invite   leave)
		case "You must join a team before readying up.":
		case "You cannot finish this early. Did you hit .done by accident?":
			// trying to finish before 5s have passed
		case "You cannot undo your finish as the race time limit has expired.":
		case "You cannot undo your finish as you have joined another race.":
		case "You cannot forfeit this early. If you are using an auto-splitter, you should configure it to not auto-reset the timer when starting a run.":
			// trying to forfeit before 5s have passed
		case "You cannot undo your forfeit as the race time limit has expired.":
		case "You cannot undo your forfeit as you have joined another race.":
		}
	}
}

type Pong struct {
	// Type string `json:"type"` //   "type": "pong"
}

func (a *App) HandlePong(data []byte) {
	var msg Pong

	err := json.Unmarshal(data, &msg)
	if err != nil {
		a.logger.Error("chat decode error: %v", err)

		return
	}

	fmt.Printf("ChatPong\n")
	fmt.Printf("[CHAT] %+v\n", msg)
}

type RaceDataMessage struct {
	Type    string    `json:"type"`
	Date    time.Time `json:"date"`
	Race    RaceData  `json:"race"`
	Version int       `json:"version"`
}

type RaceData struct {
	// Type string `json:"type"` //   "type": "race.data",
	//   "race": {
	Version               int        `json:"version"`                  // version: Integer indicating the data's version. This is incremented whenever a race changes.
	Name                  string     `json:"name"`                     // name: The race's unique name, based on the category and a randomly assigned slug.
	Category              Category   `json:"category"`                 // category: An object giving brief information about the category. Contains:
	Status                RaceStatus `json:"status"`                   // status: An object giving brief information about the race's status. Contains three keys:
	URL                   string     `json:"url"`                      // url: URL for the main race page.
	DataURL               string     `json:"data_url"`                 // data_url: URL for the race data endpoint, which you can use to obtain more detailed race information.
	WebsocketURL          string     `json:"websocket_url"`            // websocket_url: URL of the race WebSocket, used by the frontend for chat messages and real-time updates.
	WebsocketBotURL       string     `json:"websocket_bot_url"`        // websocket_bot_url: URL of the WebSocket for category bots.
	WebsocketOauthURL     string     `json:"websocket_oauth_url"`      // websocket_oauth_url: URL of the WebSocket for OAuth2-authenticated user connections. Used by third-party applications.
	Goal                  Goal       `json:"goal"`                     // goal: An object describing the race goal.
	Info                  string     `json:"info"`                     // info: String containing additional information for race entrants. This is a combination of info_bot and info_user (in that order).
	InfoBot               *string    `json:"info_bot"`                 // info_bot: String containing additional information for race entrants, as set by race bots.
	InfoUser              string     `json:"info_user"`                // info_user: String containing additional information for race entrants, as set by the monitors.
	EntrantsCount         int        `json:"entrants_count"`           // entrants_count: Total number of entrants in this race (including DQ/forfeits).
	EntrantsCountFinished int        `json:"entrants_count_finished"`  // entrants_count_finished: Total number of entrants that have finished (not counting DQ/forfeits).
	EntrantsCountInactive int        `json:"entrants_count_inactive"`  // entrants_count_inactive: Total number of entrants that have been DQed or forfieted.
	Entrants              []Entrant  `json:"entrants"`                 // entrants: The entrants list, given as an array. Ordered by race status, then by finish position (if applicable), then by score (if available), and finally by name. See below for a breakdown of entrant data blobs.
	OpenedAt              time.Time  `json:"opened_at"`                // opened_at: Date/time when the race was first created (ISO 8601 date).
	StartDelay            string     `json:"start_delay"`              // start_delay: The time allocated for the countdown, i.e. time lapse between the last entrant readying up and the race starting (ISO 8601 duration).
	StartedAt             *time.Time `json:"started_at"`               // started_at: Date/time when the race started, or null if it hasn't started yet (ISO 8601 date).
	EndedAt               *time.Time `json:"ended_at"`                 // ended_at: Date/time when the race ended, or null if it hasn't finished yet (ISO 8601 date).
	CancelledAt           *time.Time `json:"cancelled_at"`             // cancelled_at: Date/time when the race was cancelled, or null if it hasn't been cancelled (ISO 8601 date).
	Ranked                bool       `json:"ranked"`                   // ranked: Boolean indicating if the race result can be recorded when the race is concluded.
	Unlisted              bool       `json:"unlisted"`                 // unlisted: Boolean indicating an unlisted race (hidden from category view except for moderators).
	TimeLimit             string     `json:"time_limit"`               // time_limit: The maximum amount of time the race may be in progress for once it starts (ISO 8601 duration).
	TimeLimitAutoComplete bool       `json:"time_limit_auto_complete"` // time_limit_auto_complete: Boolean indicating race behaviour if the time limit is reached. If false, the race will be cancelled. If true, the race will be completed (and may still be recorded).
	RequireEvenTeams      bool       `json:"require_even_teams"`       // require_even_teams: Boolean indicating if teams must be balanced for the race to start.
	StreamingRequired     bool       `json:"streaming_required"`       // streaming_required: Boolean indiciating if entrants are required to stream in this race.
	AutoStart             bool       `json:"auto_start"`               // auto_start: Boolean indicating if the race will start automatically when all entrants are ready.
	OpenedBy              User       `json:"opened_by"`                // opened_by: User data blob for the user who opened the race room, or null if the room was opened by a bot. If present, this user is always a race monitor.
	Monitors              []User     `json:"monitors"`                 // monitors: Array of user data blobs for race monitors (in addition to the room opener) in this race.
	Recordable            bool       `json:"recordable"`               // recordable: Boolean indicating a race can be recorded once it's finished. A moderator may still opt to not record the race.
	Recorded              bool       `json:"recorded"`                 // recorded: Boolean indicating if the race has been recorded by a moderator.
	RecordedBy            *User      `json:"recorded_by"`              // recorded_by: User data blob of the moderator who recorded this race.
	DisqualifyUnready     bool       `json:"disqualify_unready"`       // disqualify_unready: Boolean indicating if users will be disqualified if they are entered into the race but do not ready up (only applies to 1v1 ladder races)
	AllowComments         bool       `json:"allow_comments"`           // allow_comments: Boolean indicating if users may add a glib remark after they finish racing.
	HideComments          bool       `json:"hide_comments"`            // hide_comments: Boolean indicating if entrant comments will be hidden until the race is finished (or cancelled).
	HideEntrants          bool       `json:"hide_entrants"`            // hide_entrants: Boolean indiciating if entrant identities are currently anonymised.
	ChatRestricted        bool       `json:"chat_restricted"`          // chat_restricted: Boolean indicating if chat restrictions are currently in place (due to allow_prerace_chat or other settings).
	AllowPreraceChat      bool       `json:"allow_prerace_chat"`       // allow_prerace_chat: Boolean indicating if users may chat while the race is preparing (does not affect monitors or moderators).
	AllowMidraceChat      bool       `json:"allow_midrace_chat"`       // allow_midrace_chat: Boolean indicating if users may chat while the race is in progress (does not affect monitors or moderators).
	AllowNonEntrantChat   bool       `json:"allow_non_entrant_chat"`   // allow_non_entrant_chat: Boolean indicating if users who have not entered the race may chat while the race is in progress (does not affect moderators).
	ChatMessageDelay      string     `json:"chat_message_delay"`       // chat_message_delay: Length of time where chat messages will only appear for race monitors (ISO 8601 duration).
	// bot_meta: Object containing custom data (see the setmeta command for further details).
}

// This will be received:
// 1) On first connect
// 2) After a system message
// 3) After a getrace action
func (a *App) HandleRaceData(data []byte) {
	var msg RaceDataMessage

	err := json.Unmarshal(data, &msg)
	if err != nil {
		a.logger.Error("chat decode error: %v", err)

		return
	}

	fmt.Printf("[RACE] %+v\n", msg)

	race := msg.Race

	a.logger.Info(
		"race update status=%s entrants=%d finished=%d",
		race.Status.State,
		race.EntrantsCount,
		race.EntrantsCountFinished,
	)

	previousStatus := a.CurrentRace.Status
	a.CurrentRace.Status = race.Status.State

	delay := time.Duration(0)
	// Parse ISO-8601 duration like PT10S
	if race.StartDelay != "" {
		dur, err := duration.FromString(race.StartDelay)
		if err != nil {
			log.Println("failed to parse start_delay:", err)
		} else {
			delay = dur.ToDuration()
		}
	}
	a.CurrentRace.Delay = delay.Milliseconds()

	a.CurrentRace.Version = msg.Version
	a.CurrentRace.Goal = race.Goal.Name
	a.CurrentRace.Info = race.Info
	a.CurrentRace.Game = race.Category.Name
	a.CurrentRace.RaceID = race.Category.Slug

	a.CurrentRace.EntrantCount = race.EntrantsCount
	a.CurrentRace.EntrantFinishedCount = race.EntrantsCountFinished
	a.CurrentRace.EntrantInactiveCount = race.EntrantsCountInactive

	a.CurrentRace.Ranked = race.Ranked
	a.CurrentRace.AutoStart = race.AutoStart
	a.CurrentRace.StatusVerbose = race.Status.VerboseValue
	a.CurrentRace.StatusHelpText = race.Status.HelpText
	a.CurrentRace.DisqualifyUnready = race.DisqualifyUnready

	if !a.CurrentRace.DisplayResults {
		for i := range race.Entrants {
			race.Entrants[i].Status.FinishTime = nil
			race.Entrants[i].Status.FinishedAt = nil
			race.Entrants[i].Place = nil
			race.Entrants[i].PlaceOrdinal = nil
			race.Entrants[i].Score = nil
			race.Entrants[i].ScoreChange = nil
			race.Entrants[i].Comment = nil
			race.Entrants[i].HasComment = nil
		}
	}

	a.CurrentRace.Entrants = race.Entrants

	if previousStatus != "in_progress" && a.CurrentRace.Status == "in_progress" {
		a.logger.Info("race transitioned to in_progress")

		if a.engine != nil {
			fmt.Println("Sending OpenSplit split command")
			a.engine.Split()
		}
	}

	// Notify frontend
	runtime.EventsEmit(a.ctx, "joinEligibility", true)
	runtime.EventsEmit(a.ctx, "raceStateUpdated", a.CurrentRace)

	a.Send(MessageDataEnvelope{
		Action: "gethistory",
		Data: MessageData{
			GUID: a.generateGUID(),
		},
	})
}

func (a *App) HandleRenders(data []byte) {
	// We don't use this, but it removes an "error"
}

// true for forfeit; false for unforfeit
func (a *App) Forfeit(state bool) {
	fmt.Printf("Forfeit status changed!")
	// if forfeited unforfeit otherwise forfeit
	a.engine.CLEAR_RUNTIME_OFFSET()
	if state {
		a.engine.Done()
		a.SendText(".forfeit", a.generateGUID())
	} else {
		a.engine.UnDone()
		a.SendText(".unforfeit", a.generateGUID())
	}
}

// true for done; false for undone
func (a *App) Done(state bool) {
	fmt.Printf("Done status changed!")
	a.engine.CLEAR_RUNTIME_OFFSET()
	if state {
		a.engine.Done()
		a.SendText(".done", a.generateGUID())
	} else {
		a.engine.UnDone()
		a.SendText(".undone", a.generateGUID())
	}
}

// true for ready; false for unready
func (a *App) Ready(state bool) {
	fmt.Printf("Ready status changed!")
	a.engine.SET_RUNTIME_OFFSET(a.CurrentRace.Delay)
	if state {
		a.SendText(".ready", a.generateGUID())
	} else {
		a.SendText(".unready", a.generateGUID())
	}
}

// true for join; false for leave
func (a *App) Join(state bool) {
	fmt.Printf("Join status changed!")
	if state {
		a.engine.SET_RUNTIME_OFFSET(a.CurrentRace.Delay)
		a.SendText(".join", a.generateGUID())
	} else {
		a.engine.CLEAR_RUNTIME_OFFSET()
		a.SendText(".leave", a.generateGUID())
	}
}

// true for hide results; false for show results
func (a *App) HideResults(state bool) {
	a.CurrentRace.DisplayResults = state

	if !state {
		a.Send(MessageDataEnvelope{
			Action: "getrace",
			Data: MessageData{
				GUID: a.generateGUID(),
			},
		})
	} else {
		for i := range a.CurrentRace.Entrants {
			a.CurrentRace.Entrants[i].Status.FinishTime = nil
			a.CurrentRace.Entrants[i].Status.FinishedAt = nil
			a.CurrentRace.Entrants[i].Place = nil
			a.CurrentRace.Entrants[i].PlaceOrdinal = nil
			a.CurrentRace.Entrants[i].Score = nil
			a.CurrentRace.Entrants[i].ScoreChange = nil
			a.CurrentRace.Entrants[i].Comment = nil
			a.CurrentRace.Entrants[i].HasComment = nil
		}

		// Notify frontend
		runtime.EventsEmit(a.ctx, "entrantsUpdated", a.CurrentRace.Entrants)
	}
}

func (a *App) SaveLog() {
	// save chat box text to file
	a.SendText(".log", a.generateGUID())
}

// func (a *App) ForceReload() {
// 	// no idea
// }

// open websocket connection and start goroutines
func (a *App) WebSocketConnection(raceURL string) error {
	// should probably do this with authorization header as shown here:
	// https://github.com/racetimeGG/racetime-app/wiki/Category-bots

	a.authenticatedRaceURL = socketUrl + "/ws/o/race/" + strings.Split(raceURL, "/")[2] + "?token=" + a.Token.AccessToken
	a.logger.Info("connecting websocket: %s", a.authenticatedRaceURL)

	conn, _, err := websocket.Dial(
		a.ctx,
		a.authenticatedRaceURL,
		nil,
	)

	if err != nil {
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	a.racetimeWS = conn

	go a.pingRoutine()
	go a.writeRoutine()
	go a.readRoutine()

	return nil
}

func (a *App) DisconnectRace() {
	if a.racetimeWS != nil {
		a.logger.Info("disconnecting websocket")
		a.racetimeWS.Close(websocket.StatusNormalClosure, "leaving race")
		a.racetimeWS = nil
	}
}

// Convert data to be sent to json before sending
func (a *App) Send(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	a.logger.Debug("sending websocket message: %s", string(data))

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
		if a.racetimeWS == nil {
			log.Println("websocket connection is nil")
			return
		}

		select {
		case msg := <-a.writeCh:
			a.logger.Debug("ws write: %s", string(msg))

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
		if a.racetimeWS == nil {
			log.Println("websocket connection is nil")
			return
		}

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

		a.logger.Debug("ws read raw: %s", string(data))
		a.logger.Debug("ws message type: %s", base.Type)

		handler, ok := a.handlers[base.Type]
		if !ok {
			a.logger.Warn("unknown websocket message type: %s", base.Type)
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
			if a.racetimeWS == nil {
				log.Println("websocket connection is nil")
				return
			}

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
	a.logger.Info(
		"token refreshed access_expiry=%s refresh_present=%v",
		a.Token.Expiry.Format(time.RFC3339),
		a.Token.RefreshToken != "",
	)
	a.logger.Debug("Access token: %s\n", a.Token.AccessToken)
	a.logger.Debug("Refresh token: %s\n", a.Token.RefreshToken)
	a.logger.Debug("Token type: %s\n", a.Token.TokenType)
	a.logger.Debug("Access token expires: %v\n", a.Token.ExpiresIn)

	securestore.SaveToken("token.enc", *a.Token, a.encryptionKey)
}

// Can only be done if the user is logged in. Refreshes tokens that needs to be stored.
// Example response should include: access_token, refresh_token, token_type, expires_in, scope
func (a *App) refreshTokens() error {
	ctx := context.Background()

	ts := a.conf.TokenSource(ctx, a.Token)

	tok, err := ts.Token()
	if err != nil {
		log.Println("refresh failed:", err)

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

	fmt.Printf("Access token: %s\n", a.Token.AccessToken)
	fmt.Printf("Refresh token: %s\n", a.Token.RefreshToken)
	fmt.Printf("Token type: %s\n", a.Token.TokenType)
	fmt.Printf("Access token expires: %s\n", a.Token.Expiry)
	fmt.Printf("Access token expires: %v\n", a.Token.ExpiresIn)

	err = securestore.SaveToken("token.enc", *a.Token, a.encryptionKey)
	if err != nil {
		log.Println("failed to save refreshed token:", err)
		return err
	}

	return nil
}

func (a *App) CheckTokens() string {
	log.Println("CheckTokens called")

	if a.Token == nil {
		log.Println("Token is nil")
		return ""
	}

	if a.Token.Valid() {
		log.Println("Token valid")
		return a.getAccessToken()
	}

	log.Println("Token expired, refreshing")

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

func (a *App) getAccessToken() (accessToken string) {
	go a.getUserInfo()
	return a.Token.AccessToken
}

func (a *App) invalidateAuth(reason string) {
	log.Println("authentication invalid:", reason)

	a.Token = nil

	// Remove encrypted token file
	err := securestore.DeleteToken("token.enc")
	if err != nil {
		log.Println("failed to delete token:", err)
	}

	// Notify frontend
	runtime.EventsEmit(a.ctx, "authExpired", reason)

	// Disconnect websocket if connected
	a.DisconnectRace()
}

func (a *App) getUserInfo() {
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(a.Token))

	resp, err := client.Get(WebRaceServer + "/o/userinfo")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		a.invalidateAuth("Authentication expired.")
		return
	}

	// if resp.StatusCode != http.StatusOK {
	// 	fmt.Println("status:", resp.Status)
	// 	fmt.Println(string(body))
	// 	return
	// }

	var user UserInfo

	if err := json.Unmarshal(body, &user); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", user)

	a.User = user

	runtime.EventsEmit(a.ctx, "userInfo", a.User)
}
