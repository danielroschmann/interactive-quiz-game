// Package game works as the game engine
package game

import (
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"os"
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

type RoundType string

const (
	RoundTypeMultipleChoice RoundType = "multiple_choice"
	RoundTypeFastestBuzz    RoundType = "fastest_buzz"
	RoundTypeChallenge      RoundType = "challenge"
)

type Game struct {
	Players         []Player
	Teams           []Team
	Rounds          []Round
	State           GameState
	CurrentRound    int
	CurrentQuestion int
	nextPlayerID    uint
}

type Round struct {
	RoundNumber int        `json:"round_number"`
	Type        RoundType  `json:"type"`
	Questions   []Question `json:"questions"`
}

type Question struct {
	Question      string   `json:"question"`
	AnswerOptions []string `json:"answer_options"`
	CorrectAnswer string   `json:"correct_answer"`
}

func NewGame(rounds []Round) *Game {
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

func (g *Game) createTeams() {
	team1 := NewTeam(1, "Team 1")
	team2 := NewTeam(2, "Team 2")

	g.Teams = []Team{team1, team2}
}

func (g *Game) assignPlayersToTeams() {
	amountOfPlayers := len(g.Players)
	rand.Shuffle(amountOfPlayers, func(i, j int) {
		g.Players[i], g.Players[j] = g.Players[j], g.Players[i]
	})

	divideTeams := amountOfPlayers / 2

	teamOneID := g.Teams[0].TeamID
	teamTwoID := g.Teams[1].TeamID

	teamOnePlayers := g.Players[:divideTeams]
	g.Teams[0].PlayerIDs = make([]uint, len(teamOnePlayers))

	for i := 0; i < len(teamOnePlayers); i++ {
		teamOnePlayers[i].TeamID = &teamOneID
		g.Teams[0].PlayerIDs[i] = teamOnePlayers[i].PlayerID
	}

	teamTwoPlayers := g.Players[divideTeams:]
	g.Teams[1].PlayerIDs = make([]uint, len(teamTwoPlayers))

	for i := 0; i < len(teamTwoPlayers); i++ {
		teamTwoPlayers[i].TeamID = &teamTwoID
		g.Teams[1].PlayerIDs[i] = teamTwoPlayers[i].PlayerID
	}
}

func LoadRounds(fileName string) ([]Round, error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	questions, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var rounds []Round

	err = json.Unmarshal(questions, &rounds)
	if err != nil {
		return nil, err
	}

	return rounds, nil
}

func (g *Game) StartGame() error {
	if len(g.Players) < 2 {
		g.State = GameStateLobby
		return ErrNotEnoughPlayers
	}

	g.createTeams()

	g.assignPlayersToTeams()

	g.CurrentRound = 1
	g.CurrentQuestion = 0
	g.State = GameStateInProgress

	return nil
}
