package game

import (
	"github.com/notnil/chess"
)

// type Command interface {
// 	isCommand()
// }

type CommandCreateGame struct {
}

type CommandJoinAdmin struct {
	Player     Player
	AdminToken AdminToken
	ReplayC    chan MessageCommandReply
}

// func (CommandJoinAdmin) isCommand() {}

type CommandJoinPlayer struct {
	Player  Player
	ReplayC chan MessageCommandReply
}

// func (CommandJoinPlayer) isCommand() {}

type CommandPlayMove struct {
	Move Move
	// ReplayC chan MessageCommandReply
}

// func (CommandPlayMove) isCommand() {}

type CommandStatus uint

const (
	CommandNotAllowed CommandStatus = iota
	CommandOK
	CommandFailed
	CommandMoveNotAllowed
)

// func (MessageCommand) isMessage() {}

type MessageCommandReply struct {
	Status  CommandStatus
	Payload interface{}
}

type CommandGetMoves struct {
	Player Player
}

func (CommandGetMoves) isCommand() {}

type ContinueWithGameResponse struct {
	PossibleMoves []chess.Move  `json:"possible_moves"`
	Outcome       chess.Outcome `json:"outcome"`
	Method        chess.Method  `json:"method"`
}
