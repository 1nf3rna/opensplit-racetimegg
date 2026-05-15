package processing

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Command byte

const (
	QUIT Command = iota
	NEW
	LOAD
	EDIT
	CANCEL
	SUBMIT
	CLOSE
	RESET
	SAVE
	SPLIT
	UNDO
	SKIP
	PAUSE
	TOGGLEGLOBAL
	FOCUS
	HELLO
)

type Engine struct {
	// L                    *lua.LState
	m sync.Mutex
	// values               map[string]emulator.Value
	conn                 net.PacketConn
	osAddr               *net.UDPAddr
	openSplitConnected   bool
	opensplitConnectedCh chan bool
	// tickFunc             *lua.LFunction
}

func NewEngine() (*Engine, chan bool) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		panic(err)
	}

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:6767")
	if err != nil {
		panic(err)
	}

	e := &Engine{
		m: sync.Mutex{},
		// values:               make(map[string]emulator.Value),
		conn:                 conn,
		osAddr:               addr,
		opensplitConnectedCh: make(chan bool),
	}

	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			e.openSplitConnected = e.Hello()
			e.updateConnectionStatus(e.openSplitConnected)
		}
	}()

	return e, e.opensplitConnectedCh
}

func (e *Engine) Close() {
	// if e.L != nil {
	// e.L.Close()
	// }
	e.updateConnectionStatus(false)
	_ = e.conn.Close()
	e.conn = nil
}

func (e *Engine) OpenSplitConnected() bool {
	return e.openSplitConnected
}

func (e *Engine) UnSplit() bool {
	packet := buildRCPacket(UNDO, false)

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		fmt.Println(err)
		// e.updateConnectionStatus(false)
		return false
	}

	// e.updateConnectionStatus(true)
	return true
}

func (e *Engine) Split() bool {
	packet := buildRCPacket(SPLIT, false)

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		fmt.Println(err)
		// e.updateConnectionStatus(false)
		return false
	}

	// e.updateConnectionStatus(true)
	return true
}

func (e *Engine) Hello() bool {
	packet := buildRCPacket(HELLO, true)

	e.m.Lock()
	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		e.m.Unlock()
		fmt.Println(err)
		return false
	}

	buf := make([]byte, 7)
	_ = e.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, _, err = e.conn.ReadFrom(buf)
	if err != nil || buf[6] != 0 {
		e.m.Unlock()
		fmt.Println(err)
		return false
	}
	e.m.Unlock()
	return true
}

func buildRCPacket(command Command, requestAck bool) []byte {
	var payload = make([]byte, 7)
	payload[0] = 'O' //magic
	payload[1] = 'S'
	payload[2] = 'R'
	payload[3] = 'C'
	payload[4] = 1 // version
	if requestAck {
		payload[5] = 1
	} else {
		payload[5] = 0
	}
	payload[6] = byte(command)

	return payload
}

func (e *Engine) updateConnectionStatus(status bool) {
	e.openSplitConnected = status
	select {
	case e.opensplitConnectedCh <- e.openSplitConnected:
	default:
	}
}
