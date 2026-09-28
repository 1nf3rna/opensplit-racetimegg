package processing

import (
	"encoding/binary"
	"fmt"
	"net"
	"opensplit-racetimegg/command"
	"opensplit-racetimegg/logger"
	"sync"
	"time"
)

var log = logger.Module("processing/engine").SetLevel(logger.DebugLevel)

const (
	RecordCommand = 0x00
	RecordAck     = 0x80
)

// type Event struct {
// 	Command Command
// }

type Engine struct {
	m                    sync.Mutex
	conn                 net.PacketConn
	osAddr               *net.UDPAddr
	openSplitConnected   bool
	opensplitConnectedCh chan bool
	events               chan command.Command
	lastHelloAck         time.Time
	done                 chan struct{}
}

func NewEngine() (*Engine, chan bool, error) {
	log.Info("creating engine")

	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		log.Error("failed to create UDP listener: %v", err)
		return nil, nil, err
	}

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:6768")
	if err != nil {
		log.Error("failed to resolve OpenSplit address: %v", err)
		return nil, nil, err
	}

	e := &Engine{
		m:                    sync.Mutex{},
		conn:                 conn,
		osAddr:               addr,
		opensplitConnectedCh: make(chan bool),
		events:               make(chan command.Command, 32),
		done:                 make(chan struct{}),
	}

	log.Info("starting read loop")

	go e.readLoop()

	go func() {
		log.Debug("starting HELLO heartbeat loop")

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				e.Hello()

				e.m.Lock()
				lastAck := e.lastHelloAck
				e.m.Unlock()

				connected := time.Since(lastAck) < 3*time.Second
				e.updateConnectionStatus(connected)

			case <-e.done:
				log.Info("heartbeat loop stopped")
				return
			}
		}
	}()

	return e, e.opensplitConnectedCh, nil
}

func (e *Engine) readLoop() {
	log.Info("read loop started")

	buf := make([]byte, 1024)

	for {
		select {
		case <-e.done:
			log.Info("read loop stopped")
			return
		default:
		}

		e.m.Lock()
		conn := e.conn
		e.m.Unlock()

		if conn == nil {
			log.Warn("connection is nil, exiting read loop")
			return
		}

		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}

			log.Error("ReadFrom failed: %v", err)
			continue
		}

		log.Debug("received packet (%d bytes)", n)

		if n < 7 {
			log.Warn("received undersized packet (%d bytes)", n)
			continue
		}

		if buf[0] != 'O' || buf[1] != 'S' || buf[2] != 'R' || buf[3] != 'C' {
			log.Warn("received packet with invalid magic bytes")
			continue
		}

		switch buf[5] {
		case RecordAck:
			status := buf[6]

			log.Debug("received ACK status=%d", status)

			if status == 0 {
				e.m.Lock()
				e.lastHelloAck = time.Now()
				e.m.Unlock()
			}

		case RecordCommand:
			cmd := command.Command(buf[6])

			log.Debug("received command=%s", commandName(cmd))

			switch cmd {
			case command.HELLO:
				e.m.Lock()
				e.lastHelloAck = time.Now()
				e.m.Unlock()

				log.Debug("received HELLO ACK")

			case command.DONE:
				log.Debug("queueing DONE event")
				select {
				case e.events <- command.DONE:
				default:
					log.Warn("event queue full, dropping DONE event")
				}

			case command.UNDONE:
				log.Debug("queueing UNDONE event")
				select {
				case e.events <- command.UNDONE:
				default:
					log.Warn("event queue full, dropping UNDONE event")
				}

			case command.PAUSE:
				log.Debug("queueing PAUSE event")
				select {
				case e.events <- command.PAUSE:
				default:
					log.Warn("event queue full, dropping PAUSE event")
				}
			}
		default:
			log.Warn("unknown record type=%d", buf[5])
		}
	}
}

func (e *Engine) Close() {
	log.Info("closing engine")

	close(e.done) // 👈 signal all goroutines to stop

	e.updateConnectionStatus(false)

	e.m.Lock()
	conn := e.conn
	e.conn = nil
	e.m.Unlock()

	if conn != nil {
		err := conn.Close()
		if err != nil {
			log.Error("failed to close UDP connection: %v", err)
		}
	}

	log.Info("engine closed")
}

func (e *Engine) Events() <-chan command.Command {
	return e.events
}

func (e *Engine) OpenSplitConnected() bool {
	return e.openSplitConnected
}

func (e *Engine) SET_RUNTIME_OFFSET(delay int64) bool {
	packet := buildRCPacket(command.SET_RUNTIME_OFFSET, &delay, false)

	log.Debug(
		"sending command=%s delay=%d",
		commandName(command.SET_RUNTIME_OFFSET),
		delay,
	)

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		log.Error("WriteTo failed: %v", err)
		return false
	}

	return true
}

func (e *Engine) CLEAR_RUNTIME_OFFSET() bool {
	packet := buildRCPacket(command.CLEAR_RUNTIME_OFFSET, nil, false)

	log.Debug("sending command=%s", commandName(command.CLEAR_RUNTIME_OFFSET))

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		log.Error("WriteTo failed: %v", err)
		return false
	}

	return true
}

func (e *Engine) SetForfeitPaused(paused bool) bool {
	state := int64(0)
	if paused {
		state = 1
	}
	return e.send(
		buildRCPacket(command.PAUSE, &state, false),
		fmt.Sprintf("PAUSE (forfeit=%t)", paused),
	)
}

func (e *Engine) send(packet []byte, name string) bool {
	log.Debug("sending command=%s", name)
	e.m.Lock()
	defer e.m.Unlock()
	if e.conn == nil {
		return false
	}
	if _, err := e.conn.WriteTo(packet, e.osAddr); err != nil {
		log.Error("WriteTo failed: %v", err)
		return false
	}
	return true
}

func (e *Engine) UnDone() bool {
	packet := buildRCPacket(command.UNDONE, nil, false)

	log.Debug("sending command=%s", commandName(command.UNDONE))

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		log.Error("WriteTo failed: %v", err)
		return false
	}

	return true
}

func (e *Engine) Done() bool {
	packet := buildRCPacket(command.DONE, nil, false)

	log.Debug("sending command=%s", commandName(command.DONE))

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		log.Error("WriteTo failed: %v", err)
		return false
	}

	return true
}

func (e *Engine) Split() bool {
	packet := buildRCPacket(command.SPLIT, nil, false)

	log.Debug("sending command=%s", commandName(command.SPLIT))

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		log.Error("WriteTo failed: %v", err)
		return false
	}

	return true
}

func (e *Engine) Hello() bool {
	packet := buildRCPacket(command.HELLO, nil, true)

	log.Debug("sending command=%s", commandName(command.HELLO))

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		log.Error("WriteTo failed: %v", err)
		return false
	}

	return true
}

func buildRCPacket(command command.Command, payload *int64, requestAck bool) []byte {
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

		log.Debug("encoded payload=%d bytes=%v", *payload, bs)

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
		log.Info("connection status changed connected=%v", status)
	}

	e.openSplitConnected = status

	select {
	case e.opensplitConnectedCh <- e.openSplitConnected:
	default:
		log.Warn("connection status channel full, dropping update")
	}
}

func commandName(cmd command.Command) string {
	switch cmd {
	case command.HELLO:
		return "HELLO"
	case command.DONE:
		return "DONE"
	case command.UNDONE:
		return "UNDONE"
	case command.SPLIT:
		return "SPLIT"
	case command.SET_RUNTIME_OFFSET:
		return "SET_RUNTIME_OFFSET"
	case command.CLEAR_RUNTIME_OFFSET:
		return "CLEAR_RUNTIME_OFFSET"
	case command.PAUSE:
		return "PAUSE"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", cmd)
	}
}
