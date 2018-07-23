package server

import (
	"chess/game"
	"chess/game/client"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// func TestServer(t *testing.T) {

// 	timeout := 1 * time.Second

// 	srv := New("game1", timeout)
// 	cl1 := client.New(srv.C)
// 	fmt.Println(cl1)

// 	// testTimeout := time.NewTimer(timeout + 1)
// 	// go func() {
// 	// 	<-testTimeout
// 	// 	t.Fatal("test timeout")
// 	// }

// 	srv.Start()

// 	//testTimeout.Stop()

// }

// func TestServer_TimeoutBeforeAdminJoined(t *testing.T) {

// 	timeout := 1 * time.Second
// 	timerBeforeTimeout := time.NewTimer(timeout / 2)
// 	timerAfterTimeout := time.NewTimer(timeout * 2)

// 	srv := New("game1", timeout)
// 	srv.Start()

// 	<-timerBeforeTimeout.C

// 	assert.NotEmpty(t, srv.state)

// 	<-timerAfterTimeout.C

// 	assert.Empty(t, srv.state)

// }

// func TestServer_TimeoutBeforePlayerTwoJoined(t *testing.T) {

// 	timeout := 1 * time.Second
// 	timerBeforeTimeout := time.NewTimer(timeout / 2)
// 	timerAfterTimeout := time.NewTimer(timeout * 2)

// 	srv := New("game1", timeout)
// 	srv.Start()

// 	replies := make(chan game.MessageCommandReply)

// 	<-timerBeforeTimeout.C
// 	srv.C <- game.CommandJoinAdmin{
// 		Command: game.Command{
// 			Reply: replies,
// 		},
// 		Player:     game.Player{Name: "name", Color: game.PieceColor(1)},
// 		AdminToken: srv.AdminToken,
// 	}

// 	reply := <-replies
// 	assert.NotEmpty(t, reply)
// 	assert.Equal(t, reply.Status, game.CommandOK)
// 	assert.NotEmpty(t, srv.state)

// 	fmt.Printf("%+v", reply)
// 	fmt.Printf("%+v", srv)

// 	<-timerAfterTimeout.C

// 	assert.Empty(t, srv.state)
// }

// func TestServer_TimeoutPlayerTwoJoinedWithSameColor(t *testing.T) {

// 	timeout := 1 * time.Second
// 	timerBeforeTimeout := time.NewTimer(timeout / 2)

// 	srv := New("game1", timeout)
// 	srv.Start()

// 	replies := make(chan game.MessageCommandReply)

// 	<-timerBeforeTimeout.C
// 	srv.C <- game.CommandJoinAdmin{
// 		Command: game.Command{
// 			Reply: replies,
// 		},
// 		Player:     game.Player{Name: "Player1", Color: game.PieceColor(1)},
// 		AdminToken: srv.AdminToken,
// 	}
// 	fmt.Println("srvasd", srv)
// 	assert.Equal(t, "poor", "rich")
// 	reply := <-replies
// 	assert.Equal(t, reply.Status, game.CommandOK)

// 	srv.C <- game.CommandJoinPlayer{
// 		Command: game.Command{
// 			Reply: replies,
// 		},
// 		Player: game.Player{Name: "Player2", Color: game.PieceColor(2)},
// 	}

// 	reply = <-replies
// 	assert.Equal(t, reply.Status, game.CommandFailed)

// 	srv.C <- game.CommandJoinPlayer{
// 		Command: game.Command{
// 			Reply: replies,
// 		},
// 		Player: game.Player{Name: "Player2", Color: game.PieceColor(2)},
// 	}

// 	// todo: join player with good color value

// 	reply = <-replies
// 	assert.Equal(t, reply.Status, game.CommandOK)

// 	assert.Equal(t, srv.PlayerBlack.Color, game.PieceColor(1))
// 	assert.Equal(t, srv.PlayerWhite.Color, game.PieceColor(2))

// }

// func TestServer_Gameplay(t *testing.T) {

// 	srv, replies := newInitializedServer(t)

// 	// Should be good.
// 	srv.C <- game.CommandPlayMove{
// 		Command: game.Command{
// 			Reply: replies,
// 		},
// 		Move: "a2a3",
// 	}

// 	reply := <-replies
// 	assert.Equal(t, reply.Status, game.CommandOK)

// 	// Should also be good.
// 	srv.C <- game.CommandPlayMove{
// 		Command: game.Command{
// 			Reply: replies,
// 		},
// 		Move: "a7a6",
// 	}

// 	reply = <-replies
// 	assert.Equal(t, reply.Status, game.CommandOK)

// 	// Should fail!
// 	srv.C <- game.CommandPlayMove{
// 		Command: game.Command{
// 			Reply: replies,
// 		},
// 		Move: "b7a8",
// 	}

// 	reply = <-replies
// 	assert.Equal(t, reply.Status, game.CommandMoveNotAllowed)

// 	fmt.Printf("%+v", srv.GameState.FEN())

// }

// func newInitializedServer(t *testing.T) (*Server, chan game.MessageCommandReply) {
// 	srv := New("game1", 1000*time.Second)
// 	srv.Start()

// 	var reply game.MessageCommandReply
// 	replies := make(chan game.MessageCommandReply)

// 	srv.C <- game.CommandJoinAdmin{
// 		Command: game.Command{
// 			Reply: replies,
// 		},
// 		Player:     game.Player{Name: "Player1", Color: game.PieceColor(1)},
// 		AdminToken: srv.AdminToken,
// 	}

// 	reply = <-replies
// 	assert.Equal(t, reply.Status, game.CommandOK)

// 	srv.C <- game.CommandJoinPlayer{
// 		Command: game.Command{
// 			Reply: replies,
// 		},
// 		Player: game.Player{Name: "Player2", Color: game.PieceColor(2)},
// 	}

// 	reply = <-replies
// 	assert.Equal(t, reply.Status, game.CommandOK)

// 	return srv, replies
// }

// func TestServer_TimeoutWaitingForAdmin(t *testing.T) {

// 	//Client to server command channel
// 	serverC := make(game.MessageChannel)

// 	//server timeout
// 	timeout := 1 * time.Second
// 	testTimeout := time.NewTimer(2 * time.Second)
// 	//new client
// 	//cl := client.New(serverC)

// 	srv := New("game1", timeout, serverC)
// 	srv.Run()

// 	<-testTimeout.C
// 	resp := <-srv.B
// 	assert.Equal(t, resp.Type, "game.timeout")
// }

// func TestServer_TimeoutWaitngForPlayerTwo(t *testing.T) {
// 	//Client to server command channel
// 	serverC := make(game.MessageChannel)

// 	//server timeout
// 	timeout := 10 * time.Second
// 	timerBeforeTimeout := time.NewTimer(timeout / 2)
// 	//new client
// 	cl := client.New(serverC)

// 	srv := New("game1", timeout, serverC)
// 	srv.Start()

// 	//definition of Player
// 	playerC := make(chan interface{})
// 	replayC := make(chan game.MessageCommandReply)
// 	player := game.Player{Name: "Daki", Color: game.PieceColor(1), C: playerC, ReplayToC: replayC}

// 	<-timerBeforeTimeout.C
// 	res, _ := cl.JoinAdmin(player, game.AdminToken([]byte("123")))
// 	assert.Equal(t, res, nil)
// }

// func TestServer_TimeoutWaitingForMove(t *testing.T) {
// 	//Client to server command channel
// 	serverC := make(game.MessageChannel)

// 	//server timeout
// 	timeout := 5 * time.Second
// 	timerBeforeTimeout := time.NewTimer(timeout / 2)
// 	//new client
// 	cl := client.New(serverC)

// 	srv := New("game1", timeout, serverC)
// 	srv.Start()

// 	//definition of Player
// 	player1C := make(chan interface{})
// 	replay1C := make(chan game.MessageCommandReply)

// 	player2C := make(chan interface{})
// 	replay2C := make(chan game.MessageCommandReply)

// 	player1 := game.Player{Name: "Daki", Color: game.PieceColor(1), C: player1C, ReplayToC: replay1C}
// 	player2 := game.Player{Name: "Daki2", Color: game.PieceColor(1), C: player2C, ReplayToC: replay2C}

// 	<-timerBeforeTimeout.C
// 	res, err := cl.JoinAdmin(player1, game.AdminToken([]byte("123")))
// 	fmt.Println("res1", res)
// 	fmt.Println("err", err)
// 	res2, err := cl.JoinPlayer(player2)
// 	fmt.Println("res2", res2)
// 	fmt.Println("err", err)
// 	fmt.Println("srv", srv)

// 	rsp := <-srv.B
// 	fmt.Println(rsp)
// 	assert.Equal(t, "res", "nil")
// }

func TestServer_GamePlay(t *testing.T) {

	//Client to server command channel
	serverC := make(game.MessageChannel)

	player1C := make(chan interface{})
	replay1C := make(chan game.MessageCommandReply)

	player2C := make(chan interface{})
	replay2C := make(chan game.MessageCommandReply)

	//server timeout
	timeout := 5 * time.Second
	timerBeforeTimeout := time.NewTimer(timeout / 2)

	//definition of Player
	player1 := game.Player{Name: "Daki", Color: game.PieceColor(1), C: player1C, ReplayToC: replay1C}  //black
	player2 := game.Player{Name: "Daki2", Color: game.PieceColor(2), C: player2C, ReplayToC: replay2C} //white

	//new client
	cl1 := client.New(serverC, player1)
	cl2 := client.New(serverC, player2)
	srv := New("game1", timeout, serverC)
	srv.Start()

	<-timerBeforeTimeout.C
	res, err := cl1.JoinAdmin(game.AdminToken([]byte("123")))
	fmt.Println("res1", res)
	fmt.Println("err", err)
	res2, err := cl2.JoinPlayer()
	fmt.Println("res2", res2)
	fmt.Println("err", err)
	fmt.Println("srv", srv)

	err, as := cl2.MakeMove("a2a3")
	fmt.Println("err", err)
	fmt.Println("as", as)

	fmt.Println("fen", srv.FEN)

	err, as = cl1.MakeMove("a7a6")
	fmt.Println("err", err)
	fmt.Println("as", as)

	fmt.Println("fen", srv.FEN)
	assert.Equal(t, "res", "nil")

}
