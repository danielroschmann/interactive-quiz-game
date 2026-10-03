// Package game works as the game engine
package game

import (
	"errors"
	"slices"
)

var ErrNotEnoughPlayers = errors.New("not enough players to start the game")

type JoinGamePayload struct {
	PlayerName string `json:"player_name"`
}

type PlayerJoinedPayload struct {
	PlayerID   uint   `json:"player_id"`
	PlayerName string `json:"player_name"`
}

type PlayerLeftPayload struct {
	PlayerID   uint   `json:"player_id"`
	PlayerName string `json:"player_name"`
}
type GameState string

const (
	GameStateLobby      GameState = "lobby"
	GameStateInProgress GameState = "in_progress"
	GameStateFinished   GameState = "finished"
)

type Game struct {
	Players      []Player
	Teams        []Team
	State        GameState
	CurrentRound int
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

func (g *Game) LeaveGame(playerID uint) (Player, bool) {
	for _, p := range g.Players {
		if p.PlayerID == playerID {
			g.Players = slices.DeleteFunc(g.Players, func(p Player) bool {
				return p.PlayerID == playerID
			})
			return p, true
		}
	}
	return Player{}, false
}

func (g *Game) CreateTeams() []Team {
	teams := []Team{}
	for i := 0; i < 2; i++ {
		teams = append(teams, NewTeam())
	}
	return teams
}

func (g *Game) StartGame() error {
	if len(g.Players) < 2 {
		g.State = GameStateLobby
		return ErrNotEnoughPlayers
	}
	g.State = GameStateInProgress

	return nil
}
