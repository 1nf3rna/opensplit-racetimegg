package processing

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"opensplit-racetimegg/logging"
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
	DONE
	UNDONE
	SET_RUNTIME_OFFSET
	CLEAR_RUNTIME_OFFSET
)

type Event struct {
	Command Command
}

const component = "ENGINE"

var logger = logging.NewLogger(true)

type Engine struct {
	m                    sync.Mutex
	conn                 net.PacketConn
	osAddr               *net.UDPAddr
	openSplitConnected   bool
	opensplitConnectedCh chan bool
	events               chan Event
	lastHelloAck         time.Time
}

func NewEngine() (*Engine, chan bool) {
	logger.Info(component, "creating engine")

	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		logger.Error(component, "failed to create UDP listener: %v", err)
		panic(err)
	}

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:6767")
	if err != nil {
		logger.Error(component, "failed to resolve OpenSplit address: %v", err)
		panic(err)
	}

	e := &Engine{
		m:                    sync.Mutex{},
		conn:                 conn,
		osAddr:               addr,
		opensplitConnectedCh: make(chan bool),
		events:               make(chan Event, 32),
	}

	logger.Info(component, "starting read loop")

	go e.readLoop()

	go func() {
		logger.Debug(component, "starting HELLO heartbeat loop")

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			e.Hello()

			connected := time.Since(e.lastHelloAck) < 3*time.Second
			e.updateConnectionStatus(connected)
		}
	}()

	return e, e.opensplitConnectedCh
}

func (e *Engine) readLoop() {
	logger.Info(component, "read loop started")

	buf := make([]byte, 1024)

	for {
		if e.conn == nil {
			logger.Warn(component, "connection is nil, exiting read loop")
			return
		}

		_ = e.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

		n, _, err := e.conn.ReadFrom(buf)
		if err != nil {
			// timeout
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}

			logger.Error(component, "ReadFrom failed: %v", err)
			continue
		}

		logger.Debug(component, "received packet (%d bytes)", n)

		// minimum packet size
		if n < 7 {
			logger.Warn(component, "received undersized packet (%d bytes)", n)
			continue
		}

		// magic
		if buf[0] != 'O' ||
			buf[1] != 'S' ||
			buf[2] != 'R' ||
			buf[3] != 'C' {
			logger.Warn(component, "received packet with invalid magic bytes")
			continue
		}

		cmd := Command(buf[6])

		logger.Info(component, "received command=%s", commandName(cmd))

		switch cmd {

		// HELLO ACK
		case HELLO:
			e.lastHelloAck = time.Now()

			logger.Debug(component, "received HELLO ACK")

		// OpenSplit emitted DONE
		case DONE:
			logger.Debug(component, "queueing DONE event")

			select {
			case e.events <- Event{Command: DONE}:
			default:
				logger.Warn(component, "event queue full, dropping DONE event")
			}

		// OpenSplit emitted UNDONE
		case UNDONE:
			logger.Debug(component, "queueing UNDONE event")

			select {
			case e.events <- Event{Command: UNDONE}:
			default:
				logger.Warn(component, "event queue full, dropping UNDONE event")
			}
		}
	}
}

func (e *Engine) Close() {
	logger.Info(component, "closing engine")

	e.updateConnectionStatus(false)

	if e.conn != nil {
		err := e.conn.Close()
		if err != nil {
			logger.Error(component, "failed to close UDP connection: %v", err)
		}
	}

	e.conn = nil

	logger.Info(component, "engine closed")
}

func (e *Engine) Events() <-chan Event {
	return e.events
}

func (e *Engine) OpenSplitConnected() bool {
	return e.openSplitConnected
}

func (e *Engine) SET_RUNTIME_OFFSET(delay int64) bool {
	packet := buildRCPacket(SET_RUNTIME_OFFSET, &delay, false)

	logger.Info(
		component,
		"sending command=%s delay=%d",
		commandName(SET_RUNTIME_OFFSET),
		delay,
	)

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		logger.Error(component, "WriteTo failed: %v", err)
		return false
	}

	return true
}

func (e *Engine) CLEAR_RUNTIME_OFFSET() bool {
	payload := int64(0)

	packet := buildRCPacket(CLEAR_RUNTIME_OFFSET, &payload, false)

	logger.Info(component, "sending command=%s", commandName(CLEAR_RUNTIME_OFFSET))

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		logger.Error(component, "WriteTo failed: %v", err)
		return false
	}

	return true
}

func (e *Engine) UnDone() bool {
	packet := buildRCPacket(UNDONE, nil, false)

	logger.Info(component, "sending command=%s", commandName(UNDONE))

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		logger.Error(component, "WriteTo failed: %v", err)
		return false
	}

	return true
}

func (e *Engine) Done() bool {
	packet := buildRCPacket(DONE, nil, false)

	logger.Info(component, "sending command=%s", commandName(DONE))

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		logger.Error(component, "WriteTo failed: %v", err)
		return false
	}

	return true
}

func (e *Engine) Split() bool {
	packet := buildRCPacket(SPLIT, nil, false)

	logger.Info(component, "sending command=%s", commandName(SPLIT))

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		logger.Error(component, "WriteTo failed: %v", err)
		return false
	}

	return true
}

func (e *Engine) Hello() bool {
	packet := buildRCPacket(HELLO, nil, true)

	logger.Debug(component, "sending command=%s", commandName(HELLO))

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		logger.Error(component, "WriteTo failed: %v", err)
		return false
	}

	return true
}

func buildRCPacket(command Command, payload *int64, requestAck bool) []byte {
	packetSize := 7

	if payload != nil {
		packetSize += 8
	}

	packet := make([]byte, packetSize)

	packet[0] = 'O'
	packet[1] = 'S'
	packet[2] = 'R'
	packet[3] = 'C'
	packet[4] = 1

	if requestAck {
		packet[5] = 1
	} else {
		packet[5] = 0
	}

	packet[6] = byte(command)

	if payload != nil {
		bs := make([]byte, 8)

		// Little Endian (Least Significant Byte first)
		binary.LittleEndian.PutUint64(bs, uint64(*payload))

		logger.Debug(component, "encoded payload=%d bytes=%v", *payload, bs)

		packet[7] = bs[0]
		packet[8] = bs[1]
		packet[9] = bs[2]
		packet[10] = bs[3]
		packet[11] = bs[4]
		packet[12] = bs[5]
		packet[13] = bs[6]
		packet[14] = bs[7]
	}

	return packet
}

func (e *Engine) updateConnectionStatus(status bool) {
	if e.openSplitConnected != status {
		logger.Info(component, "connection status changed connected=%v", status)
	}

	e.openSplitConnected = status

	select {
	case e.opensplitConnectedCh <- e.openSplitConnected:
	default:
		logger.Warn(component, "connection status channel full, dropping update")
	}
}

func commandName(cmd Command) string {
	switch cmd {
	case HELLO:
		return "HELLO"
	case DONE:
		return "DONE"
	case UNDONE:
		return "UNDONE"
	case SPLIT:
		return "SPLIT"
	case SET_RUNTIME_OFFSET:
		return "SET_RUNTIME_OFFSET"
	case CLEAR_RUNTIME_OFFSET:
		return "CLEAR_RUNTIME_OFFSET"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", cmd)
	}
}
