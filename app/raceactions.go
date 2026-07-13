package app

import (
	"strings"

	"opensplit-racetimegg/racetime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type entrantState struct {
	Status         string
	Joined         bool
	RaceLocked     bool
	HasTwitch      bool
	StreamLive     bool
	StreamOverride bool
}

func (a *App) updateRaceActions() {
	actions := a.calculateRaceActions()
	runtime.EventsEmit(a.ctx, "raceActionsUpdated", actions)
}

func (a *App) calculateRaceActions() racetime.RaceActions {
	state := a.currentEntrantState()

	actions := racetime.RaceActions{}

	a.calculateJoinActions(&actions, state)
	a.calculateReadyActions(&actions, state)
	a.calculateFinishActions(&actions, state)

	return actions
}

func (a *App) currentEntrantState() entrantState {
	state := entrantState{
		Status:    "not_joined",
		HasTwitch: strings.TrimSpace(a.User.TwitchName) != "",
		RaceLocked: a.CurrentRace.Status == "in_progress" ||
			a.CurrentRace.EndedAt != nil ||
			a.CurrentRace.CancelledAt != nil,
	}

	for i := range a.CurrentRace.Entrants {
		e := &a.CurrentRace.Entrants[i]

		if e.User.Id != a.User.ID {
			continue
		}

		state.Status = e.Status.Value
		state.StreamLive = e.StreamLive
		state.StreamOverride = e.StreamOverride

		break
	}

	switch state.Status {
	case "ready",
		"not_ready",
		"in_progress",
		"done",
		"dnf",
		"dq":
		state.Joined = true
	}

	return state
}

func (a *App) calculateJoinActions(
	actions *racetime.RaceActions,
	state entrantState,
) {
	actions.JoinAction = racetime.JoinActionJoin
	actions.LeaveAction = racetime.LeaveActionLeave

	switch {
	case state.RaceLocked:
		actions.JoinReason = "Race unavailable"

	case !a.canJoin:
		actions.JoinReason = "Not eligible"

	case a.CurrentRace.StreamingRequired && !state.HasTwitch:
		actions.JoinReason = "Link Twitch account"

	case a.CurrentRace.Status == "invitational":
		switch state.Status {
		case "invited":
			actions.CanJoin = true
			actions.JoinAction = racetime.JoinActionAcceptInvite

			actions.CanLeave = true
			actions.LeaveAction = racetime.LeaveActionDeclineInvite

		case "requested":
			actions.CanLeave = true
			actions.LeaveAction = racetime.LeaveActionCancelInvite

		default:
			if !state.Joined {
				actions.CanJoin = true
				actions.JoinAction = racetime.JoinActionRequestInvite
			}

			if state.Joined {
				actions.CanLeave = true
				actions.LeaveAction = racetime.LeaveActionLeave
			}
		}

	default:
		if state.Joined {
			actions.CanLeave = true
		} else {
			actions.CanJoin = true
		}
	}
}

func (a *App) calculateReadyActions(
	actions *racetime.RaceActions,
	state entrantState,
) {
	switch {
	case state.RaceLocked:
		actions.ReadyReason = "Race unavailable"

	case !state.Joined:
		actions.ReadyReason = "Join race first"

	case a.CurrentRace.StreamingRequired:
		switch {
		case !state.HasTwitch:
			actions.ReadyReason = "Link Twitch account"

		case !(state.StreamLive || state.StreamOverride):
			actions.StreamBlocked = true
			actions.ReadyReason = "Stream must be live"

		default:
			actions.CanReady = true
		}

	default:
		actions.CanReady = true
	}
}

func (a *App) calculateFinishActions(
	actions *racetime.RaceActions,
	state entrantState,
) {
	running := a.CurrentRace.Status == "in_progress"

	actions.CanDone = state.Joined && running
	actions.CanForfeit = state.Joined && running

	if !actions.CanDone {
		actions.DoneReason = "Race not in progress"
	}

	if !actions.CanForfeit {
		actions.ForfeitReason = "Race not in progress"
	}
}
