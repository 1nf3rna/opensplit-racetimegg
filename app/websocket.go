package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"opensplit-racetimegg/racetime"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) websocketURL(raceURL string) string {
	slug := strings.Split(raceURL, "/")[2]

	return fmt.Sprintf(
		"%s/ws/o/race/%s?token=%s",
		racetime.SocketURL,
		slug,
		a.Token.AccessToken,
	)
}

// open websocket connection and start goroutines
func (a *App) ConnectRace(raceURL string) error {
	if err := a.DisconnectRace(); err != nil {
		return err
	}

	// should probably do this with authorization header as shown here:
	// https://github.com/racetimeGG/racetime-app/wiki/Category-bots
	url := a.websocketURL(raceURL)

	log.Info("connecting websocket: %s", url)

	a.CurrentRace = racetime.RaceInfo{}
	runtime.EventsEmit(a.ctx, "raceUpdated", a.CurrentRace)
	runtime.EventsEmit(a.ctx, "chatUpdated", []racetime.ChatMessage{})

	ws, _, err := websocket.Dial(context.Context(a.ctx), url, nil)
	if err != nil {
		return err
	}

	ws.SetReadLimit(8 * 1024 * 1024)

	conn := NewRaceConnection(a.ctx, ws)

	a.raceConn = conn

	conn.wg.Add(3)

	go func() {
		defer conn.wg.Done()
		a.readRoutine(conn)
	}()

	go func() {
		defer conn.wg.Done()
		a.writeRoutine(conn)
	}()

	go func() {
		defer conn.wg.Done()
		a.pingRoutine(conn)
	}()

	return nil
}

func (a *App) DisconnectRace() error {
	if a.raceConn != nil {
		log.Info("disconnecting websocket")

		a.raceConn.Close()
		a.raceConn = nil
	}

	a.CurrentRace = racetime.RaceInfo{}
	a.canJoin = false

	runtime.EventsEmit(a.ctx, "raceUpdated", a.CurrentRace)
	runtime.EventsEmit(a.ctx, "chatUpdated", []racetime.ChatMessage{})
	runtime.EventsEmit(a.ctx, "raceActions", racetime.RaceActions{})

	return nil
}

func (a *App) sendAction(action string) {
	a.SendText(action, a.generateGUID())
}

// Convert data to be sent to json before sending
func (a *App) Send(v any) error {
	if a.raceConn == nil {
		return errors.New("not connected")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	log.Debug("sending websocket message: %s", string(data))

	select {
	case a.raceConn.writeCh <- data:
		return nil

	case <-a.raceConn.ctx.Done():
		return errors.New("connection closed")
	}
}

// message writing routine
func (a *App) writeRoutine(conn *RaceConnection) {
	for {
		select {

		case msg := <-conn.writeCh:
			log.Debug("ws write: %s", string(msg))

			ctx, cancel := context.WithTimeout(conn.ctx, 5*time.Second)

			err := conn.ws.Write(
				ctx,
				websocket.MessageText,
				msg,
			)

			cancel()

			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Debug("write routine stopped")
					return
				}

				switch websocket.CloseStatus(err) {
				case websocket.StatusNormalClosure:
					log.Info("websocket closed normally")
				default:
					log.Error("write error: %v", err)
				}
				return
			}

		case <-conn.ctx.Done():
			return
		}
	}
}

// message recieve and routing routine
func (a *App) readRoutine(conn *RaceConnection) {
	for {
		_, data, err := conn.ws.Read(conn.ctx)

		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Debug("read routine stopped")
				return
			}

			switch websocket.CloseStatus(err) {
			case websocket.StatusNormalClosure:
				log.Info("websocket closed normally")
			default:
				log.Error("read error: %v", err)
			}
			return
		}

		var base racetime.BaseMessage

		err = json.Unmarshal(data, &base)
		if err != nil {
			log.Error("invalid json:", err)
			continue
		}

		log.Debug("ws read raw: %s", string(data))
		log.Debug("ws message type: %s", base.Type)

		handler, ok := a.handlers[base.Type]
		if !ok {
			log.Warn("unknown websocket message type: %s", base.Type)
			continue
		}

		if conn != a.raceConn {
			return
		}

		handler(data)
	}
}

// keep alive goroutine
func (a *App) pingRoutine(conn *RaceConnection) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Timeout for ping response
			ctx, cancel := context.WithTimeout(conn.ctx, 5*time.Second)

			err := conn.ws.Ping(ctx)

			cancel()

			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Debug("ping routine stopped")
					return
				}

				switch websocket.CloseStatus(err) {
				case websocket.StatusNormalClosure:
					log.Info("websocket closed normally")
				default:
					log.Error("ping failed:", err)
				}
				return
			}

			log.Info("ping sent")
		case <-conn.ctx.Done():
			return
		}
	}
}

func (a *App) SendText(text string, GUID string) {
	log.Debug("SendText called")

	a.Send(racetime.MessageDataEnvelope{
		Action: "message",
		Data: racetime.MessageData{
			Message: text,
			Pinned:  false,
			GUID:    GUID,
		},
	})
}
