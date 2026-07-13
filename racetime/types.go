package racetime

import (
	"encoding/json"
	"fmt"
	"time"
)

type RaceActions struct {
	CanJoin    bool       `json:"canJoin"`
	JoinReason string     `json:"joinReason"`
	JoinAction JoinAction `json:"joinAction"`

	CanLeave    bool        `json:"canLeave"`
	LeaveReason string      `json:"leaveReason"`
	LeaveAction LeaveAction `json:"leaveAction"`

	CanReady    bool   `json:"canReady"`
	ReadyReason string `json:"readyReason"`

	CanDone    bool   `json:"canDone"`
	DoneReason string `json:"doneReason"`

	CanForfeit    bool   `json:"canForfeit"`
	ForfeitReason string `json:"forfeitReason"`

	StreamBlocked bool `json:"streamBlocked"`
}

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
	UserID               string
	Version              int
	Goal                 string
	Game                 string
	RaceID               string
	Info                 string
	StreamingRequired    bool
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
	EndedAt              *time.Time
	CancelledAt          *time.Time
}

type Entrant struct {
	User           User          `json:"user"` // user: User data blob for this entrant.
	Status         EntrantStatus `json:"status"`
	FinishTime     *string       `json:"finish_time"`     // finish_time: The user's final finish time, or null if they've not finished (ISO 8601 duration).
	FinishedAt     *time.Time    `json:"finished_at"`     // finished_at: The date/time when the user finished, or null if they've not finished (ISO 8601 date).
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
	Id            string `json:"id"`             // "id": "fR42gLweew3pQlm4",
	FullName      string `json:"full_name"`      // "full_name": "Mario#5527",
	Name          string `json:"name"`           // "name": "Mario",
	Discriminator string `json:"discriminator"`  // "discriminator": "5527",
	Url           string `json:"url"`            // "url": "/user/fR42gLweew3pQlm4",
	Avatar        string `json:"avatar"`         // "avatar": "/media/mario.png",
	Pronouns      string `json:"pronouns"`       // "pronouns": "he/him",
	Flair         string `json:"flair"`          // "flair": "monitor supporter",
	TwitchName    string `json:"twitch_name"`    // "twitch_name": "ItsaMeMario",
	TwitchChannel string `json:"twitch_channel"` // "twitch_channel": "https://www.twitch.tv/itsamemario",
	CanModerate   bool   `json:"can_moderate"`   // "can_moderate": false
}

type ConnectionState struct {
	ConnectionStatus ConnectionStatus `json:"connection_status"`
	Message          string           `json:"message"`
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
