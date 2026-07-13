package app

import (
	"context"
	"sync"

	"github.com/coder/websocket"
)

type RaceConnection struct {
	ws *websocket.Conn

	ctx    context.Context
	cancel context.CancelFunc

	writeCh chan []byte

	once sync.Once
	wg   sync.WaitGroup
}

func NewRaceConnection(parent context.Context, ws *websocket.Conn) *RaceConnection {
	ctx, cancel := context.WithCancel(parent)

	return &RaceConnection{
		ws: ws,

		ctx:    ctx,
		cancel: cancel,

		writeCh: make(chan []byte, 64),
	}
}

func (c *RaceConnection) Close() {
	c.once.Do(func() {
		c.cancel()

		c.wg.Wait()

		_ = c.ws.Close(websocket.StatusNormalClosure, "disconnect")
	})
}
