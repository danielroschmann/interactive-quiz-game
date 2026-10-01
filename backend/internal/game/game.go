// Package game works as the game engine
package game

type JoinGamePayload struct {
	PlayerName string `json:"player_name"`
}

type PlayerJoinedPayload struct {
	PlayerID   uint   `json:"player_id"`
	PlayerName string `json:"player_name"`
}

type Game struct {
	Players      []Player
	nextPlayerID uint
}

func NewGame() *Game {
	return &Game{
		Players:      make([]Player, 0),
		nextPlayerID: 1,
	}
}

func (g *Game) JoinGame(playerName string) Player {
	player := Player{
		PlayerName: playerName,
		Points:     0,
		PlayerID:   g.nextPlayerID,
	}
	g.Players = append(g.Players, player)
	g.nextPlayerID++
	return player
}
