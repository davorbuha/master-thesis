package server

import (
	"chess/errors"
	"chess/game"
	"chess/subpub"
	"fmt"
	"time"

	"github.com/notnil/chess"
)

var uniqueGameId game.GameID = 0

//timeout const
const defaultSetupTimeout = 5 * time.Minute

// func (p PlayerWithMessageChannel) Notify(event subpub.Event) {
// 	p.C <- event
// }

type Server struct {
	ID           game.GameID
	Name         string
	B            game.BroadcastChannel // broadcast channel
	C            game.MessageChannel   // one channel for players or spectators to join! (command channel)
	AdminToken   game.AdminToken
	PlayerWhite  game.Player
	PlayerBlack  game.Player
	GameState    *chess.Game
	moveTimeout  time.Duration
	setupTimeout time.Duration
	state        stateFn
	StateStr     string
	FEN          string
	Started      bool
	subpub       *subpub.SubPub
	AdminColor   string
}

func New(name string, timeout time.Duration, serverC game.MessageChannel, adminColor string) *Server {
	uniqueGameId++

	return &Server{
		ID:           uniqueGameId,
		Name:         name,
		B:            make(game.BroadcastChannel, 100),
		C:            serverC,
		AdminToken:   game.RandStringRunes(8),
		GameState:    chess.NewGame(chess.UseNotation(chess.LongAlgebraicNotation{})),
		moveTimeout:  timeout,
		setupTimeout: defaultSetupTimeout,
		Started:      false,
		AdminColor:   adminColor,
	}
}

func (srv *Server) setGameState(state string) {
	srv.StateStr = state
	srv.FEN = srv.GameState.FEN()

	srv.B <- game.BroadcastMessage{
		Type: "game.update",
		Payload: struct {
			State string
			FEN   string
		}{
			srv.StateStr,
			srv.FEN,
		},
	}
}

type stateFn func() stateFn

func (srv *Server) waitForBlackMove() stateFn {
	outcome := srv.GameState.Outcome()
	if outcome != chess.NoOutcome {
		return nil
	}
	srv.setGameState("waitForBlackMove")
	timer := time.NewTimer(srv.moveTimeout)
	for {
		select {
		case <-timer.C:
			//broadcast to subs that game is finished
			return nil

		case msg := <-srv.PlayerBlack.C:
			cmd, ok := msg.(game.CommandPlayMove)
			if !ok {
				srv.PlayerBlack.ReplayToC <- game.MessageCommandReply{
					Status: game.CommandNotAllowed,
				}
				continue
			}

			// Try to apply desired move, and fail gracefully.
			err := srv.GameState.MoveStr(string(cmd.Move))
			if err != nil {
				srv.PlayerBlack.ReplayToC <- game.MessageCommandReply{
					Status:  game.CommandMoveNotAllowed,
					Payload: err,
				}
				continue
			}

			// Send command success.

			srv.PlayerBlack.ReplayToC <- game.MessageCommandReply{
				Status:  game.CommandOK,
				Payload: true,
			}

			// Broadcast to subs played move and possible next moves.

			return srv.waitForWhiteMove
		}
	}
}

func (srv *Server) waitForWhiteMove() stateFn {
	outcome := srv.GameState.Outcome()
	if outcome != chess.NoOutcome {
		return nil
	}
	srv.setGameState("waitForWhiteMove")

	timer := time.NewTimer(srv.moveTimeout)
	for {
		select {
		case <-timer.C:
			//broadcast to subs that game is finished
			return nil

		case msg := <-srv.PlayerWhite.C:
			cmd, ok := msg.(game.CommandPlayMove)
			if !ok {
				srv.PlayerWhite.ReplayToC <- game.MessageCommandReply{
					Status: game.CommandNotAllowed,
				}
				continue
			}

			// Try to apply desired move, and fail gracefully.
			err := srv.GameState.MoveStr(string(cmd.Move))
			if err != nil {
				srv.PlayerWhite.ReplayToC <- game.MessageCommandReply{
					Status:  game.CommandMoveNotAllowed,
					Payload: err,
				}
				continue
			}

			// Send command success.
			srv.PlayerWhite.ReplayToC <- game.MessageCommandReply{
				Status:  game.CommandOK,
				Payload: true,
			}
			return srv.waitForBlackMove
		}
	}
}

func (srv *Server) startWithGame() stateFn {
	srv.StateStr = "startWithGame"
	srv.Started = true
	srv.B <- game.BroadcastMessage{
		Type: "game.started",
	}

	return srv.waitForWhiteMove
}

func (srv *Server) JoinPlayerTwo(player game.Player) error {
	if srv.PlayerBlack.C == nil {
		player.Color = 0
		srv.PlayerBlack = player
		return nil
	}
	if srv.PlayerWhite.C == nil {
		player.Color = 1
		srv.PlayerWhite = player
		return nil
	}
	return errors.New("Unable to join")
}

func (srv *Server) waitForPlayerTwo() stateFn {
	srv.StateStr = "waitForPlayerTwo"
	// timer := time.NewTimer(srv.setupTimeout)

	for {
		select {
		// case <-timer.C:
		// 	srv.B <- game.BroadcastMessage{Type: "game.timeout"}
		// 	return nil

		case msg := <-srv.C:
			cmd, ok := msg.(game.CommandJoinPlayer)
			if !ok {
				cmd.ReplayC <- game.MessageCommandReply{
					Status: game.CommandNotAllowed,
				}
				continue
			}

			err := srv.JoinPlayerTwo(cmd.Player)
			if err != nil {
				cmd.ReplayC <- game.MessageCommandReply{
					Status: game.CommandFailed,
				}
				continue
			}
			fmt.Println("srvvv", srv)
			cmd.ReplayC <- game.MessageCommandReply{
				Status: game.CommandOK,
			}

			return srv.startWithGame
		}
	}
}

func (srv *Server) JoinAdmin(player game.Player, token game.AdminToken) error {

	if srv.AdminToken != token {
		return game.ErrWrongAdminToken
	}

	if player.Color == game.PieceColor(0) {
		srv.PlayerBlack = player
		return nil
	} else {
		srv.PlayerWhite = player
		return nil
	}

	//After success joining broadcast to all players that admin joined and subscribe admin
	// srv.subpub.Broadcast(EventAdminJoined{player})
	// srv.subpub.Subcribe(srv.PlayerWhite)

	return errors.New("Unable to join")
}

func (srv *Server) waitForAdmin() stateFn {
	srv.StateStr = "waitForAdmin"
	timer := time.NewTimer(srv.setupTimeout)

	for {
		select {
		case <-timer.C:
			srv.B <- game.BroadcastMessage{Type: "game.timeout"}
			return nil

		case msg := <-srv.C:
			cmd, ok := msg.(game.CommandJoinAdmin)
			if !ok {
				cmd.ReplayC <- game.MessageCommandReply{
					Status: game.CommandNotAllowed,
				}
				break
			}

			err := srv.JoinAdmin(cmd.Player, cmd.AdminToken)
			if err != nil {
				cmd.ReplayC <- game.MessageCommandReply{
					Status:  game.CommandFailed,
					Payload: err,
				}
				break
			}

			cmd.ReplayC <- game.MessageCommandReply{
				Status: game.CommandOK,
			}

			return srv.waitForPlayerTwo
		}
	}

}

func (srv *Server) init() {
	srv.StateStr = "init"
	srv.state = srv.waitForAdmin
	for {
		srv.state = srv.state()
		if srv.state == nil {
			return
		}
	}
}

func (srv *Server) Run() {
	srv.init()
}

func (srv *Server) Start() {
	go srv.init()
}

func (srv *Server) Dispose() {
	// Dispose of stuff at the end-of-life of the Server.
	close(srv.B)
	close(srv.C)

	if srv.PlayerWhite.C != nil {
		close(srv.PlayerWhite.C)
		close(srv.PlayerWhite.ReplayToC)
	}

	if srv.PlayerBlack.C != nil {
		close(srv.PlayerBlack.C)
		close(srv.PlayerBlack.ReplayToC)
	}
}
