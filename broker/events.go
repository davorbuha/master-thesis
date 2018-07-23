package broker

import (
	"chess/game"
	"chess/game/server"
)

type GameEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func buildGameCreatedEvent(g *server.Server) GameEvent {
	return GameEvent{
		Type: "game.created",
		Payload: struct {
			ID         game.GameID `json:"id"`
			Name       string      `json:"name"`
			AdminColor string      `json:"admin_color"`
		}{
			g.ID,
			g.Name,
			g.AdminColor,
		},
	}
}

func buildGameStartedEvent(g *server.Server) GameEvent {
	return GameEvent{
		Type: "game.started",
		Payload: struct {
			ID game.GameID `json:"id"`
		}{
			g.ID,
		},
	}
}

func buildGameUpdatedEvent(g *server.Server) GameEvent {
	return GameEvent{
		Type: "game.updated",
		Payload: struct {
			ID    game.GameID `json:"id"`
			State string      `json:"state"`
			FEN   string      `json:"fen"`
		}{
			g.ID,
			g.StateStr,
			g.FEN,
		},
	}
}

func buildGameClosedEvent(g *server.Server) GameEvent {
	return GameEvent{
		Type: "game.closed",
		Payload: struct {
			ID game.GameID `json:"id"`
		}{
			g.ID,
		},
	}
}
