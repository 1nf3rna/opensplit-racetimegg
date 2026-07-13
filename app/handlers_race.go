package app

import (
	"encoding/json"
	"opensplit-racetimegg/racetime"
	"time"

	duration "github.com/channelmeter/iso8601duration"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// This will be received:
// 1) On first connect
// 2) After a system message
// 3) After a getrace action
func (a *App) HandleRaceData(data []byte) {
	var msg racetime.RaceDataMessage

	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Error("chat decode error: %v", err)

		return
	}

	log.Info("[RACE] %+v\n", msg)

	a.CurrentRace.UserID = a.User.ID

	race := msg.Race

	log.Info(
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
			log.Error("failed to parse start_delay:", err)
		} else {
			delay = -dur.ToDuration()
		}
	}
	a.CurrentRace.Delay = delay.Milliseconds()

	a.CurrentRace.Version = msg.Version
	a.CurrentRace.Goal = race.Goal.Name
	a.CurrentRace.Info = race.Info
	a.CurrentRace.Game = race.Category.Name
	a.CurrentRace.RaceID = race.Category.Slug
	a.CurrentRace.StreamingRequired = race.StreamingRequired
	a.CurrentRace.EndedAt = race.EndedAt
	a.CurrentRace.CancelledAt = race.CancelledAt

	a.CurrentRace.EntrantCount = race.EntrantsCount
	a.CurrentRace.EntrantFinishedCount = race.EntrantsCountFinished
	a.CurrentRace.EntrantInactiveCount = race.EntrantsCountInactive

	a.CurrentRace.Ranked = race.Ranked
	a.CurrentRace.AutoStart = race.AutoStart
	a.CurrentRace.StatusVerbose = race.Status.VerboseValue
	a.CurrentRace.StatusHelpText = race.Status.HelpText
	a.CurrentRace.DisqualifyUnready = race.DisqualifyUnready

	a.CurrentRace.Entrants = race.Entrants

	if previousStatus != "in_progress" && a.CurrentRace.Status == "in_progress" {
		log.Info("race transitioned to in_progress")

		if a.engine != nil {
			log.Info("Sending OpenSplit split command")
			a.engine.Split()
		}
	}

	a.canJoin = true // optimistic until server says otherwise
	a.updateRaceActions()
	// Notify frontend
	runtime.EventsEmit(a.ctx, "raceStateUpdated", a.CurrentRace)

	a.Send(racetime.MessageDataEnvelope{
		Action: "gethistory",
		Data: racetime.MessageData{
			GUID: a.generateGUID(),
		},
	})
}
