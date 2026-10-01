package game

type JoinGamePayload struct {
	PlayerName string `json:"player_name"`
}

type Game struct {
	Players []Player
}

func NewGame() *Game {
	return &Game{
		Players: make([]Player, 0),
	}
}

func (g *Game) JoinGame(playerName string) {
	player := Player{
		PlayerName: playerName,
		Points:     0,
	}
	g.Players = append(g.Players, player)
}
