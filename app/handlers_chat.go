package app

import (
	"encoding/json"
	"opensplit-racetimegg/racetime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) emitChatUpdate() {
	msgs := append([]racetime.ChatMessage(nil), a.CurrentRace.Text...)

	runtime.EventsEmit(a.ctx, "chatUpdated", msgs)
}

func (a *App) HandleChatMessage(data []byte) {
	var env racetime.ChatMessageEnvelope

	err := json.Unmarshal(data, &env)
	if err != nil {
		log.Error("chat decode error: %v", err)
		return
	}

	msg := env.Message

	log.Info("ChatMessage\n")
	log.Info("[CHAT] %+v\n", msg)

	// ignore duplicate messages
	for _, m := range a.CurrentRace.Text {
		if m.ID == msg.ID {
			return
		}
	}

	a.CurrentRace.Text = append(a.CurrentRace.Text, msg)

	// Notify frontend
	a.emitChatUpdate()
}

func (a *App) HandleChatHistory(data []byte) {
	var msg racetime.ChatHistory

	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Error("chat decode error: %v", err)

		return
	}

	log.Info("ChatHistory\n")
	log.Info("[CHAT] %+v\n", msg)

	messages := make([]racetime.ChatMessage, len(msg.Messages))
	copy(messages, msg.Messages)

	// // replace race message array
	a.CurrentRace.Text = messages

	// // Notify frontend
	a.emitChatUpdate()
}

func (a *App) HandleChatDM(data []byte) {
	var msg racetime.ChatDM

	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Error("chat decode error: %v", err)

		return
	}

	log.Info("ChatDM\n")
	log.Info("[CHAT] %+v\n", msg)

	// This message type doesn't matter
}

func (a *App) setPinned(id string, pinned bool) {
	for i := range a.CurrentRace.Text {
		if a.CurrentRace.Text[i].ID == id {
			a.CurrentRace.Text[i].IsPinned = pinned
			break
		}
	}

	a.emitChatUpdate()
}

func (a *App) HandleChatPin(data []byte) {
	var msg racetime.ChatPin

	if err := json.Unmarshal(data, &msg); err != nil {
		log.Error("chat decode error: %v", err)
		return
	}

	a.setPinned(msg.Message.ID, true)
}

func (a *App) HandleChatUnpin(data []byte) {
	var msg racetime.ChatUnpin

	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Error("chat decode error: %v", err)

		return
	}

	a.setPinned(msg.Message.ID, false)
}

func (a *App) HandleChatDelete(data []byte) {
	var env racetime.ChatDeleteEnvelope

	err := json.Unmarshal(data, &env)
	if err != nil {
		log.Error("chat decode error: %v", err)

		return
	}

	msg := env.Delete

	log.Info("ChatDelete\n")
	log.Info("[CHAT] %+v\n", msg)

	for i, m := range a.CurrentRace.Text {
		if m.ID == msg.ID {
			// Remove element at index i
			a.CurrentRace.Text = append(a.CurrentRace.Text[:i], a.CurrentRace.Text[i+1:]...)

			// Notify frontend
			a.emitChatUpdate()

			return
		}
	}
}

func (a *App) HandleChatPurge(data []byte) {
	var env racetime.ChatPurgeEnvelope

	err := json.Unmarshal(data, &env)
	if err != nil {
		log.Error("chat decode error: %v", err)

		return
	}

	msg := env.Purge

	log.Info("ChatPurge\n")
	log.Info("[CHAT] %+v\n", msg)

	filtered := a.CurrentRace.Text[:0]

	for _, m := range a.CurrentRace.Text {
		if m.User.Id != msg.User.Id {
			filtered = append(filtered, m)
		}
	}

	a.CurrentRace.Text = filtered

	// Notify frontend
	a.emitChatUpdate()
}
