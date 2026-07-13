package racetime

const (
	// local dev
	// socketUrl = "ws://localhost:8000"
	// WebRaceServer = "http://localhost:8000"

	// live
	SocketURL     = "wss://racetime.gg"
	WebRaceServer = "https://racetime.gg"
)

type JoinAction string

const (
	JoinActionJoin          JoinAction = "join"
	JoinActionAcceptInvite  JoinAction = "accept_invite"
	JoinActionRequestInvite JoinAction = "request_invite"
)

type LeaveAction string

const (
	LeaveActionLeave         LeaveAction = "leave"
	LeaveActionDeclineInvite LeaveAction = "decline_invite"
	LeaveActionCancelInvite  LeaveAction = "cancel_invite"
)

type ConnectionStatus byte

const (
	Disconnected ConnectionStatus = iota
	Connected
	Reconnecting
	WaitingForGame
)
