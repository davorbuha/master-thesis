package main

import (
	"chess/broker"
	"chess/game"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var brkr *broker.Broker

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	brkr = broker.New(5 * time.Minute)
	router := mux.NewRouter()
	router.HandleFunc("/create_game/{name}/move_time/{moveTime}/admin_color/{adminColor}", handleCreateGame) //at least POST
	router.HandleFunc("/close_game/{id}/admin_token/{adminToken}", handleCloseGame)
	router.HandleFunc("/game_events", handleGameEventsConnections)
	router.HandleFunc("/join_game/{id}", handleJoinGameConnections)
	http.ListenAndServe(":"+os.Getenv("PORT"), router)
}

func handleCreateGame(w http.ResponseWriter, r *http.Request) {
	mvars := mux.Vars(r)
	name := mvars["name"]
	adminColor := mvars["adminColor"]
	moveTime, err := strconv.Atoi(mvars["moveTime"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	g := brkr.NewGame(name, time.Duration(moveTime)*time.Second, adminColor)

	err = writeJsonResponse(w, struct {
		GameID     game.GameID     `json:"id"`
		AdminToken game.AdminToken `json:"admin_token"`
	}{
		GameID:     g.ID,
		AdminToken: g.AdminToken,
	})
}

func handleCloseGame(w http.ResponseWriter, r *http.Request) {
	mvars := mux.Vars(r)
	id := mvars["id"]
	adminToken := mvars["adminToken"]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = brkr.CloseGame(game.GameID(idInt), game.AdminToken(adminToken))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

}

func handleGameEventsConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("handleGameEventsConnections: %v", err)
		return
	}
	brkr.AddGameEventsConn(conn)
}

func handleJoinGameConnections(w http.ResponseWriter, r *http.Request) {
	mvars := mux.Vars(r)
	gameIdInt, err := strconv.Atoi(mvars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("handleJoinGameConnections: %v", err)
		return
	}
	err = brkr.AddGamePlayerConn(game.GameID(gameIdInt), conn)
	if err != nil {
		log.Printf("handleJoinGameConnections: %v", err)
		return
	}
}

func writeJsonResponse(w http.ResponseWriter, resp interface{}) error {
	js, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(js)
	if err != nil {
		return err
	}

	return nil
}
