package game

//TODO: Make pacakage types
type (
	PieceColor uint

	Player struct {
		Name      string
		Color     PieceColor
		C         chan interface{}
		ReplayToC chan MessageCommandReply
	}

	AdminToken string

	// MessageChannel chan MessageCommand
	MessageChannel chan interface{}
	Move           string
)

const (
	Black PieceColor = iota
	White
)

type BroadcastChannel chan BroadcastMessage
type BroadcastMessage struct {
	Type    string
	Payload interface{}
}

type GameID uint
