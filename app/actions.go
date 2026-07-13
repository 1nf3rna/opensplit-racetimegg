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
	a.engine.CLEAR_RUNTIME_OFFSET()
	if state {
		a.engine.Done()
		a.sendAction(".done")
	} else {
		a.engine.UnDone()
		a.sendAction(".undone")
	}
}

// true for forfeit; false for unforfeit
func (a *App) Forfeit(state bool) {
	log.Info("Forfeit status changed!")
	// if forfeited unforfeit otherwise forfeit
	a.engine.CLEAR_RUNTIME_OFFSET()
	if state {
		a.engine.Done()
		a.sendAction(".forfeit")
	} else {
		a.engine.UnDone()
		a.sendAction(".unforfeit")
	}
}

func (a *App) Join() {
	log.Info("Join status changed!")
	a.engine.SET_RUNTIME_OFFSET(a.CurrentRace.Delay)
	if a.CurrentRace.Status == "invitational" {
		a.sendAction(".acceptinvite")
	} else {
		a.sendAction(".join")
	}
}

func (a *App) Leave() {
	log.Info("Leaving race")
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
