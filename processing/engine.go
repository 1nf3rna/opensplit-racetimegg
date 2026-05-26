package processing

import (
	"encoding/binary"
	"fmt"
	"log"
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
	DONE
	UNDONE
	SET_RUNTIME_OFFSET
	CLEAR_RUNTIME_OFFSET
)

type Event struct {
	Command Command
}

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
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		panic(err)
	}

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:6767")
	if err != nil {
		panic(err)
	}

	e := &Engine{
		m:                    sync.Mutex{},
		conn:                 conn,
		osAddr:               addr,
		opensplitConnectedCh: make(chan bool),
		events:               make(chan Event, 32),
	}

	go e.readLoop()

	go func() {
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
	buf := make([]byte, 1024)

	for {
		if e.conn == nil {
			return
		}

		_ = e.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

		n, _, err := e.conn.ReadFrom(buf)
		if err != nil {
			// timeout
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}

			fmt.Println(err)
			continue
		}

		// minimum packet size
		if n < 7 {
			continue
		}

		// magic
		if buf[0] != 'O' ||
			buf[1] != 'S' ||
			buf[2] != 'R' ||
			buf[3] != 'C' {
			continue
		}

		cmd := Command(buf[6])
		log.Printf("[ENGINE] received command=%s", commandName(cmd))

		switch cmd {

		// HELLO ACK
		case HELLO:
			e.lastHelloAck = time.Now()

		// OpenSplit emitted DONE
		case DONE:
			select {
			case e.events <- Event{Command: DONE}:
			default:
			}

		// OpenSplit emitted UNDONE
		case UNDONE:
			select {
			case e.events <- Event{Command: UNDONE}:
			default:
			}
		}
	}
}

func (e *Engine) Close() {
	// if e.L != nil {
	// e.L.Close()
	// }
	e.updateConnectionStatus(false)
	_ = e.conn.Close()
	e.conn = nil
}

func (e *Engine) Events() <-chan Event {
	return e.events
}

func (e *Engine) OpenSplitConnected() bool {
	return e.openSplitConnected
}

func (e *Engine) SET_RUNTIME_OFFSET(delay int64) bool {
	packet := buildRCPacket(SET_RUNTIME_OFFSET, &delay, false)
	log.Printf("[ENGINE] sending command=%s", commandName(SET_RUNTIME_OFFSET))

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

func (e *Engine) CLEAR_RUNTIME_OFFSET() bool {
	payload := int64(0)
	packet := buildRCPacket(CLEAR_RUNTIME_OFFSET, &payload, false)
	log.Printf("[ENGINE] sending command=%s", commandName(CLEAR_RUNTIME_OFFSET))

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

func (e *Engine) UnDone() bool {
	packet := buildRCPacket(UNDONE, nil, false)
	log.Printf("[ENGINE] sending command=%s", commandName(UNDONE))

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

func (e *Engine) Done() bool {
	packet := buildRCPacket(DONE, nil, false)
	log.Printf("[ENGINE] sending command=%s", commandName(DONE))

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
	packet := buildRCPacket(SPLIT, nil, false)
	log.Printf("[ENGINE] sending command=%s", commandName(SPLIT))

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
	packet := buildRCPacket(HELLO, nil, true)
	log.Printf("[ENGINE] sending command=%s", commandName(HELLO))

	e.m.Lock()
	defer e.m.Unlock()

	_, err := e.conn.WriteTo(packet, e.osAddr)
	if err != nil {
		fmt.Println(err)
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

	packet[0] = 'O' //magic
	packet[1] = 'S'
	packet[2] = 'R'
	packet[3] = 'C'
	packet[4] = 1 // version
	if requestAck {
		packet[5] = 1
	} else {
		packet[5] = 0
	}
	packet[6] = byte(command)

	bs := make([]byte, 8)
	if payload != nil {
		// Little Endian (Least Significant Byte first)
		binary.LittleEndian.PutUint64(bs, uint64(*payload))
		fmt.Printf("LittleEndian: %v\n", bs)
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
	e.openSplitConnected = status
	select {
	case e.opensplitConnectedCh <- e.openSplitConnected:
	default:
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
