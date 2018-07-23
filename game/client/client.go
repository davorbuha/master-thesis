package client

import (
	"game"
	"errors"

	"github.com/gorilla/websocket"
)

type Client struct {
	Player  game.Player
	serverC game.MessageChannel
	Conn    *websocket.Conn
}

func New(player game.Player, serverC game.MessageChannel, conn *websocket.Conn) *Client {
	return &Client{
		Player:  player,
		serverC: serverC,
		Conn:    conn,
	}
}

func (c *Client) MakeMove(playersMove string) (error, interface{}) {
	c.Player.C <- game.CommandPlayMove{
		Move: game.Move(playersMove),
	}
	replay := <-c.Player.ReplayToC
	if replay.Status == game.CommandNotAllowed {
		return game.CommandNotAllowedError{Command: "CommandNotAllowed"}, nil
	}
	if replay.Status == game.CommandMoveNotAllowed {
		return errors.New("MoveNotAllowed"), nil
	}
	return nil, replay.Payload
}

func (c *Client) JoinAdmin(token game.AdminToken) (game.Player, error) {
	replyC := make(chan game.MessageCommandReply)
	c.serverC <- game.CommandJoinAdmin{
		Player:     c.Player,
		ReplayC:    replyC,
		AdminToken: token,
	}

	reply := <-replyC

	if reply.Status == game.CommandNotAllowed {
		return game.Player{}, game.CommandNotAllowedError{Command: "JoinAdmin"}
	}

	return c.Player, nil
}

func (c *Client) JoinPlayer() (game.Player, error) {
	replyC := make(chan game.MessageCommandReply)
	c.serverC <- game.CommandJoinPlayer{
		Player:  c.Player,
		ReplayC: replyC,
	}
	reply := <-replyC

	if reply.Status == game.CommandFailed {
		return game.Player{}, game.CommandNotAllowedError{Command: "JoinPlayer"}
	}

	return c.Player, nil
}
