package processing

import (
	"encoding/binary"
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
	DONE
	UNDONE
	SET_RUNTIME_OFFSET
	CLEAR_RUNTIME_OFFSET
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

func (e *Engine) SET_RUNTIME_OFFSET(delay int64) bool {
	packet := buildRCPacket(SET_RUNTIME_OFFSET, &delay, false)

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

func buildRCPacket(command Command, payload *int64, requestAck bool) []byte {
	var packet = make([]byte, 7)
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
