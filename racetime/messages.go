package racetime

import "time"

type BaseMessage struct {
	Type string `json:"type"`
}

type MessageDataEnvelope struct {
	Action string      `json:"action"`
	Data   MessageData `json:"data"`
}

type MessageData struct {
	Message  string  `json:"message,omitempty"`
	Pinned   bool    `json:"pinned,omitempty"`
	Actions  *string `json:"actions,omitempty"`
	DirectTo *string `json:"direct_to,omitempty"`
	GUID     string  `json:"guid"`
}

type ChatMessageEnvelope struct {
	Message ChatMessage `json:"message"`
}

type ChatMessage struct {
	ID           string         `json:"id"`
	User         User           `json:"user"`
	Bot          *string        `json:"bot"`
	DirectTo     *User          `json:"direct_to"`
	PostedAt     time.Time      `json:"posted_at"`
	Message      string         `json:"message"`
	MessagePlain string         `json:"message_plain"`
	Highlight    bool           `json:"highlight"`
	IsDM         bool           `json:"is_dm"`
	IsBot        bool           `json:"is_bot"`
	IsSystem     bool           `json:"is_system"`
	IsPinned     bool           `json:"is_pinned"`
	Delay        DurationString `json:"delay"`
	//	    "actions" { <action objects> }
}

type ChatHistory struct {
	Messages []ChatMessage `json:"messages"`
}

type ChatDM struct {
	Message  string `json:"message"`
	FromUser User   `json:"from_user"`
	FromBot  string `json:"from_bot"`
	To       User   `json:"to"`
}

type ChatPin struct {
	Message ChatMessage `json:"message"`
}

type ChatUnpin struct {
	Message ChatMessage `json:"message"`
}

type ChatDeleteEnvelope struct {
	Delete ChatDelete `json:"delete"`
}

type ChatDelete struct {
	ID        string  `json:"id"`
	User      *User   `json:"user"`
	Bot       *string `json:"bot"`
	IsBot     bool    `json:"is_bot"`
	DeletedBy User    `json:"deleted_by"`
}

type ChatPurgeEnvelope struct {
	Purge ChatPurge `json:"purge"`
}

type ChatPurge struct {
	User     User `json:"user"`
	PurgedBy User `json:"purged_by"`
}

type ChatError struct {
	Errors []string `json:"errors"`
}

type Pong struct{}

type RaceDataMessage struct {
	Type    string    `json:"type"`
	Date    time.Time `json:"date"`
	Race    RaceData  `json:"race"`
	Version int       `json:"version"`
}

type RaceData struct {
	Version               int        `json:"version"`
	Name                  string     `json:"name"`
	Category              Category   `json:"category"`
	Status                RaceStatus `json:"status"`
	URL                   string     `json:"url"`
	DataURL               string     `json:"data_url"`
	WebsocketURL          string     `json:"websocket_url"`
	WebsocketBotURL       string     `json:"websocket_bot_url"`
	WebsocketOauthURL     string     `json:"websocket_oauth_url"`
	Goal                  Goal       `json:"goal"`
	Info                  string     `json:"info"`
	InfoBot               *string    `json:"info_bot"`
	InfoUser              string     `json:"info_user"`
	EntrantsCount         int        `json:"entrants_count"`
	EntrantsCountFinished int        `json:"entrants_count_finished"`
	EntrantsCountInactive int        `json:"entrants_count_inactive"`
	Entrants              []Entrant  `json:"entrants"`
	OpenedAt              time.Time  `json:"opened_at"`
	StartDelay            string     `json:"start_delay"`
	StartedAt             *time.Time `json:"started_at"`
	EndedAt               *time.Time `json:"ended_at"`
	CancelledAt           *time.Time `json:"cancelled_at"`
	Ranked                bool       `json:"ranked"`
	Unlisted              bool       `json:"unlisted"`
	TimeLimit             string     `json:"time_limit"`
	TimeLimitAutoComplete bool       `json:"time_limit_auto_complete"`
	RequireEvenTeams      bool       `json:"require_even_teams"`
	StreamingRequired     bool       `json:"streaming_required"`
	AutoStart             bool       `json:"auto_start"`
	OpenedBy              User       `json:"opened_by"`
	Monitors              []User     `json:"monitors"`
	Recordable            bool       `json:"recordable"`
	Recorded              bool       `json:"recorded"`
	RecordedBy            *User      `json:"recorded_by"`
	DisqualifyUnready     bool       `json:"disqualify_unready"`
	AllowComments         bool       `json:"allow_comments"`
	HideComments          bool       `json:"hide_comments"`
	HideEntrants          bool       `json:"hide_entrants"`
	ChatRestricted        bool       `json:"chat_restricted"`
	AllowPreraceChat      bool       `json:"allow_prerace_chat"`
	AllowMidraceChat      bool       `json:"allow_midrace_chat"`
	AllowNonEntrantChat   bool       `json:"allow_non_entrant_chat"`
	ChatMessageDelay      string     `json:"chat_message_delay"`
}
