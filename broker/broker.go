package broker

import (
	"chess/game"
	"chess/game/client"
	"chess/game/server"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketsMessageType string
type WebsocketsRequest struct {
	Type    WebsocketsMessageType `json:"type"`
	Payload json.RawMessage       `json:"payload"`
}
type WebsocketsResponse struct {
	Type    WebsocketsMessageType `json:"type"`
	Payload interface{}           `json:"payload"`
}

type JoinPlayerPayload struct {
	Name  string          `json:"name"`
	Color game.PieceColor `json:"color"`
}

type JoinAdminPayload struct {
	JoinPlayerPayload
	AdminToken game.AdminToken `json:"admin_token"`
}

type MakeMovePayload struct {
	Move string `json:"move"`
}

const (
	JoinAdminMsgType  = WebsocketsMessageType("join.admin")
	JoinPlayerMsgType = WebsocketsMessageType("join.player")
	MakeMoveMsgType   = WebsocketsMessageType("game.makemove")
)

type BrokerGamesMap map[game.GameID]*server.Server
type BrokerClientsMap map[game.GameID][]*client.Client

type Broker struct {
	meout          time.Duration
	games          BrokerGamesMap
	gameClients    BrokerClientsMap
	gameEventsConn []*websocket.Conn
	locker         sync.Mutex
	createGameC    chan game.CommandCreateGame
}

func New(timeout time.Duration) *Broker {
	return &Broker{
		meout:       timeout,
		games:       BrokerGamesMap{},
		gameClients: BrokerClientsMap{},
		createGameC: make(chan game.CommandCreateGame),
	}
}

func (b *Broker) NewGame(name string, moveTimeout time.Duration, adminColor string) *server.Server {
	srvC := make(game.MessageChannel)
	srv := server.New(name, moveTimeout, srvC, adminColor)
	srv.Start()

	b.games[srv.ID] = srv
	go b.waitForBroadcast(srv)

	b.broadcastGameEvent(buildGameCreatedEvent(srv))

	return srv
}

func (b *Broker) CloseGame(id game.GameID, adminToken game.AdminToken) error {
	srv, exists := b.games[id]
	if !exists {
		return errors.New(fmt.Sprintf("Broker.CloseGame: no such game: %d", id))
	}

	if srv.AdminToken != adminToken {
		return errors.New(fmt.Sprintf("Broker.CloseGame: wrong AdminToken provided for game: ", id))
	}

	delete(b.games, id)
	srv.Dispose()

	b.broadcastGameEvent(buildGameClosedEvent(srv))

	return nil
}

func (b *Broker) broadcastGameEvent(ev GameEvent) {
	for _, conn := range b.gameEventsConn {
		err := conn.WriteJSON(ev)
		if err != nil {
			log.Printf("Broker.broadcastGameEvent: cannot write JSON: %v", err)
		}
	}
}

func (b *Broker) publishGameEvent(gid game.GameID, ev GameEvent) {
	clnts, ok := b.gameClients[gid]
	if ok {
		for _, clnt := range clnts {
			err := clnt.Conn.WriteJSON(ev)
			if err != nil {
				log.Printf("Broker.publishGameEvent: cannot write JSON: %v", err)
			}
		}
	} else {
		log.Printf("Broker.publishGameEvent: no clients for game: %d", gid)
	}
}

func (b *Broker) waitForBroadcast(srv *server.Server) {
	for {
		reply := <-srv.B
		log.Printf("waitForBroadcast: received %v", reply)

		switch reply.Type {
		case "game.timeout":
			err := b.CloseGame(srv.ID, srv.AdminToken)
			if err != nil {
				log.Printf("Broker.waitForBroadcast: cannot close game: %v", err)
			}

		case "game.started":
			ev := buildGameStartedEvent(srv)
			b.broadcastGameEvent(ev)
			b.publishGameEvent(srv.ID, ev)

		case "game.update":
			ev := buildGameUpdatedEvent(srv)
			b.publishGameEvent(srv.ID, ev)
		}
	}

	// TODO: break for loop when game is over, so that this go func stops.
}

func (b *Broker) AddGameEventsConn(conn *websocket.Conn) error {
	b.gameEventsConn = append(b.gameEventsConn, conn)

	// Replay creation events for existing games.
	log.Printf("Broker.AddGameEventsConn: found %d running games.", len(b.games))
	for id, _ := range b.games {
		if !b.games[id].Started {
			toWrite := buildGameCreatedEvent(b.games[id])
			err := conn.WriteJSON(toWrite)
			if err != nil {
				log.Printf("Broker.AddGameEventsConn: cannot write JSON: %v", err)
			}
		}
	}

	return nil
}

func (b *Broker) AddGamePlayerConn(id game.GameID, conn *websocket.Conn) error {
	srv, exists := b.games[id]
	if !exists {
		return errors.New(fmt.Sprintf("Broker.AddGamePlayerConn: no such game: %d", id))
	}

	// Create new Player for this connection.
	player := game.Player{
		C:         make(chan interface{}),
		ReplayToC: make(chan game.MessageCommandReply),
	}

	// Create Client for Player and connect it to the Server.
	clnt := client.New(player, srv.C, conn)
	b.gameClients[id] = append(b.gameClients[id], clnt)

	// Connect Client to websocket connection.
	go handleWebsocketTraffic(clnt, conn)

	return nil
}

func handleWebsocketTraffic(clnt *client.Client, conn *websocket.Conn) {
	for {
		var msg WebsocketsRequest
		err := conn.ReadJSON(&msg)
		if err != nil {
			handleWebsocketError(err, conn)
			continue
		}
		switch msg.Type {
		case JoinAdminMsgType:
			var payload JoinAdminPayload
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				fmt.Println("err", err)
				handleWebsocketError(err, conn)
				continue
			}
			clnt.Player.Name = payload.Name
			clnt.Player.Color = payload.Color
			_, err = clnt.JoinAdmin(payload.AdminToken)
			if err != nil {
				handleWebsocketError(err, conn)
				continue
			}

			err = conn.WriteJSON(true) // TODO
			if err != nil {
				handleWebsocketError(err, conn)
				continue
			}

		case JoinPlayerMsgType:
			var payload JoinPlayerPayload
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				handleWebsocketError(err, conn)
				continue
			}
			clnt.Player.Name = payload.Name
			clnt.Player.Color = payload.Color
			_, err = clnt.JoinPlayer()
			if err != nil {
				handleWebsocketError(err, conn)
			}

			err = conn.WriteJSON(true) // TODO
			if err != nil {
				handleWebsocketError(err, conn)
				continue
			}

		case MakeMoveMsgType:
			var payload MakeMovePayload
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				handleWebsocketError(err, conn)
				continue
			}

			err, _ = clnt.MakeMove(payload.Move)
			if err != nil {
				handleWebsocketError(err, conn)
				continue
			}

			err = conn.WriteJSON(true) // TODO
			if err != nil {
				handleWebsocketError(err, conn)
				continue
			}

		default:
			err := errors.New(fmt.Sprintf("Broker: unknown websockets message type: %s", msg.Type))
			handleWebsocketError(err, conn)
			continue
		}
	}
	conn.Close()
}

func handleWebsocketError(err error, conn *websocket.Conn) {
	log.Printf("%v", err)
	conn.WriteJSON(WebsocketsResponse{
		Type:    "error",
		Payload: err,
	})
}

// func (b *Broker) Notify(event pubsub.Event) {
// 	b.locker.Lock()
// 	defer b.locker.Unlock()

// 	switch event.Type {
// 	case "game.timeout":
// 		delete(b.games, ev.GameID)
// 	case "game.canceld":
// 		delete(b.games, ev.GameID)
// 	}

// 	b.publisher.Broadcast(event)
// }
