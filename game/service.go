package game

//TODO
type Service interface {
	JoinAdmin(Player, AdminToken, MessageChannel) error
	JoinPlayer(Player, MessageChannel) (Player, error)
}
