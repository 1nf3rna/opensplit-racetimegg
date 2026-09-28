package app

// true for ready; false for unready
func (a *App) Ready(state bool) {
	log.Info("Ready status changed!")
	a.engine.SET_RUNTIME_OFFSET(a.CurrentRace.Delay)
	if state {
		a.sendAction(".ready")
	} else {
		a.sendAction(".unready")
	}
}

// true for done; false for undone
func (a *App) Done(state bool) {
	log.Info("Done status changed!")
	if state {
		a.engine.Done()
	} else {
		a.engine.UnDone()
	}
	a.setRuntimeOffsetHeld("done", state)
	if state {
		a.sendAction(".done")
	} else {
		a.sendAction(".undone")
	}
}

// OpenSplitDone is called when OpenSplit completes its final split.
func (a *App) OpenSplitDone() {
	a.setRuntimeOffsetHeld("done", true)
	a.sendAction(".done")
}

// OpenSplitUndone is called when OpenSplit undoes its final split.
func (a *App) OpenSplitUndone() {
	a.setRuntimeOffsetHeld("done", false)
	a.sendAction(".undone")
}

// OpenSplitForfeit is called when OpenSplit resets the active run.
func (a *App) OpenSplitForfeit() {
	a.setRuntimeOffsetHeld("forfeit", true)
	a.sendAction(".forfeit")
}

// true for forfeit; false for unforfeit
func (a *App) Forfeit(state bool) {
	log.Info("Forfeit status changed!")
	a.setRuntimeOffsetHeld("forfeit", state)
	a.engine.SetForfeitPaused(state)
	if state {
		a.sendAction(".forfeit")
	} else {
		a.sendAction(".unforfeit")
	}
}

func (a *App) setRuntimeOffsetHeld(reason string, held bool) {
	a.timerActionMu.Lock()
	defer a.timerActionMu.Unlock()

	wasHeld := a.doneHeld || a.forfeitHeld
	switch reason {
	case "done":
		a.doneHeld = held
	case "forfeit":
		a.forfeitHeld = held
	}
	isHeld := a.doneHeld || a.forfeitHeld
	if wasHeld == isHeld {
		return
	}

	if isHeld {
		a.engine.CLEAR_RUNTIME_OFFSET()
		return
	}

	a.engine.SET_RUNTIME_OFFSET(a.CurrentRace.Delay)
}

func (a *App) resetRaceTimerStatus() {
	a.timerActionMu.Lock()
	doneHeld := a.doneHeld
	forfeitHeld := a.forfeitHeld
	a.doneHeld = false
	a.forfeitHeld = false
	a.timerActionMu.Unlock()
	if doneHeld {
		a.engine.UnDone()
	}
	if forfeitHeld {
		a.engine.SetForfeitPaused(false)
	}
}

func (a *App) Join() {
	log.Info("Join status changed!")
	a.resetRaceTimerStatus()
	a.engine.SET_RUNTIME_OFFSET(a.CurrentRace.Delay)
	if a.CurrentRace.Status == "invitational" {
		a.sendAction(".acceptinvite")
	} else {
		a.sendAction(".join")
	}
}

func (a *App) Leave() {
	log.Info("Leaving race")
	a.resetRaceTimerStatus()
	a.engine.CLEAR_RUNTIME_OFFSET()
	a.sendAction(".leave")
}

func (a *App) DeclineInvite() {
	log.Info("Declining invite")
	a.engine.CLEAR_RUNTIME_OFFSET()
	a.sendAction(".declineinvite")
}

// true for request; false for cancel
func (a *App) RequestInvite(state bool) {
	log.Info("Invite status changed!")
	if state {
		a.engine.SET_RUNTIME_OFFSET(a.CurrentRace.Delay)
		a.sendAction(".requestinvite")
	} else {
		a.engine.CLEAR_RUNTIME_OFFSET()
		a.sendAction(".cancelinvite")
	}
}

func (a *App) SaveLog() {
	// save chat box text to file
	a.sendAction(".log")
}
