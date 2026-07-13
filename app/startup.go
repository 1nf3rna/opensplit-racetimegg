package app

import (
	"context"
	"opensplit-racetimegg/processing"
	"opensplit-racetimegg/racetime"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	go func() {
		s := racetime.ConnectionState{}
		for {
			status, ok := <-a.osConnectionCh
			if !ok {
				return
			}

			s.ConnectionStatus = racetime.Disconnected
			s.Message = "OpenSplit Not Found"
			if status {
				s.ConnectionStatus = racetime.Connected
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
				log.Info("opensplit DONE event received")
				a.sendAction(".done")

			case processing.UNDONE:
				log.Info("opensplit UNDONE event received")
				a.sendAction(".undone")
			}
		}
	}()
}

func (a *App) generateGUID() string {
	return uuid.NewString()
}
