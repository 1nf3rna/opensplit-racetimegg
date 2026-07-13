package app

import (
	"encoding/json"
	"opensplit-racetimegg/racetime"
)

func (a *App) HandleChatError(data []byte) {
	var msg racetime.ChatError

	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Error("chat decode error: %v", err)

		return
	}

	log.Info("ChatError\n")
	log.Info("[CHAT] %+v\n", msg)

	// Do stuff depending on the errors
	for _, msgError := range msg.Errors {
		switch msgError {
		// Streaming required and twitch channel not linked (join   request_to_join   invite)
		case "You are not eligible to join this race.":
			log.Info("User not eligible to join race")
			// Disable join button
			a.canJoin = false
			a.updateRaceActions()
			// Inform user too many monitors
		case "Races cannot have more than 5 monitors.":
		// Inform user message is too long
		case "Ensure this value has at most 1000 characters (it has 52428).":
		// Set if race is being changed from invitational to open when not in that state
		case "Race is not an invitational.":
		// Set if race is being changed from open to invitational when not in that state
		case "Race is not open.":
		// Set trying to start race while conditions don't allow it to start (can_begin)
		case "Race cannot be started yet.":
		// Set when trying to cancel a done race
		case "Cannot cancel a race that is in %(state)s state.":
		// Set when trying to partition a race (can_partition)
		case "Race cannot be partitioned yet.":
		// Set when trying to finish a race that's not in progress (is_in_progress  finish)
		case "Cannot finish a race that has not been started.":
		//(is_unfinalized  unfinish)
		case "Cannot restart a race from this state.":
		//(hold   (un)record)
		case "Race cannot be finalized, it is on hold.":
		case "This race cannot be recorded because one or more entrants have deleted their account. Please set this race to \"Do not record\".":
		//((un)record)
		case "Race is not recordable or already recorded.":
		// (add/remove hold)
		case "Race hold cannot be changed now.":
		// Set when race in progress and room opener; can't make rematch
		case "Unable to comply, racing in progress.":
		// Set when trying to rematch when not a race monitor
		case "Only race monitors may create a rematch. Start a new race room instead.":
		// User not allowed to make races
		case "You are not allowed to start a new race.":
		// Not a team race (create team   join team   get_available_teams)
		case "Not a team race.":
		// Join race first
		case "Cannot join a team (join the race first!).":
		// Cannot join team multiple times
		case "You are already in that team.":
		case "Cannot change team during the race.":
		case "You cannot join that team without an invitation.":
		// invite or joined and disqualify_unready enabled (decline_invite   leave)
		case "You are not allowed to quit this race.":
		case "You must join a team before readying up.":
		// trying to finish before 5s have passed
		case "You cannot finish this early. Did you hit .done by accident?":
		case "You cannot undo your finish as the race time limit has expired.":
		case "You cannot undo your finish as you have joined another race.":
		// trying to forfeit before 5s have passed
		case "You cannot forfeit this early. If you are using an auto-splitter, you should configure it to not auto-reset the timer when starting a run.":
		case "You cannot undo your forfeit as the race time limit has expired.":
		case "You cannot undo your forfeit as you have joined another race.":
		}
	}
}

func (a *App) HandlePong(data []byte) {
	var raw map[string]any

	if err := json.Unmarshal(data, &raw); err != nil {
		log.Error("chat decode error: %v", err)
		return
	}

	log.Info("ChatPong\n")
}

func (a *App) HandleRenders(data []byte) {
	// We don't use this, but it removes an "error"
}
