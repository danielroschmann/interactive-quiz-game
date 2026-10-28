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

var (
	ErrNotEnoughPlayers     = errors.New("not enough players to start the game")
	ErrNoCurrentQuestion    = errors.New("no current question")
	ErrNoCurrentRound       = errors.New("no current round")
	ErrPlayerNotFound       = errors.New("player not found")
	ErrAlreadyBuzzed        = errors.New("player already buzzed")
	ErrNotBuzzedPlayer      = errors.New("player did not win the buzz")
	ErrGameNotInProgress    = errors.New("game not in progress")
	ErrWrongRoundType       = errors.New("wrong round type")
	ErrUnsupportedRoundType = errors.New("round type not supported")
	ErrInvalidRoundType     = errors.New("invalid round type")
	ErrNoCurrentChallenge   = errors.New("no current challenge")
	ErrInvalidTeamPoints    = errors.New("invalid team points")
	ErrAnswerPending        = errors.New("an answer is already pending")
)

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
	Players        []Player
	Teams          []Team
	Rounds         []Round
	State          GameState
	CurrentRound   int
	CurrentItem    int
	nextPlayerID   uint
	BuzzedPlayerID *uint
	HostPlayerID   uint
	AnswerPending  bool
}

type Round struct {
	RoundNumber int         `json:"round_number"`
	Type        RoundType   `json:"type"`
	Questions   []Question  `json:"questions,omitempty"`
	Challenges  []Challenge `json:"challenges,omitempty"`
}

type Question struct {
	Question      string   `json:"question"`
	AnswerOptions []string `json:"answer_options"`
	CorrectAnswer string   `json:"correct_answer"`
}

type Challenge struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewGame(rounds []Round) *Game {
	return &Game{
		Players:      make([]Player, 0),
		Rounds:       rounds,
		nextPlayerID: 1,
		State:        GameStateLobby,
	}
}

func (g *Game) JoinGame(playerName string) Player {
	player := Player{
		PlayerName: playerName,
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

func (g *Game) GetCurrentQuestion() (Question, error) {
	currentRound, err := g.GetCurrentRound()
	if err != nil {
		return Question{}, ErrNoCurrentQuestion
	}

	if currentRound.Type != RoundTypeMultipleChoice && currentRound.Type != RoundTypeFastestBuzz {
		return Question{}, ErrWrongRoundType
	}
	questionIndex := g.CurrentItem
	if questionIndex < 0 || questionIndex >= len(currentRound.Questions) {
		return Question{}, ErrNoCurrentQuestion
	}

	return currentRound.Questions[questionIndex], nil
}

func (g *Game) GetCurrentChallenge() (Challenge, error) {
	currentRound, err := g.GetCurrentRound()
	if err != nil {
		return Challenge{}, ErrNoCurrentChallenge
	}

	if currentRound.Type != RoundTypeChallenge {
		return Challenge{}, ErrWrongRoundType
	}

	challengeIndex := g.CurrentItem
	if challengeIndex < 0 || challengeIndex >= len(currentRound.Challenges) {
		return Challenge{}, ErrNoCurrentChallenge
	}
	return currentRound.Challenges[challengeIndex], nil
}

func (g *Game) GetCurrentRound() (Round, error) {
	roundIndex := g.CurrentRound - 1
	if roundIndex < 0 || roundIndex >= len(g.Rounds) {
		return Round{}, ErrNoCurrentRound
	}
	return g.Rounds[roundIndex], nil
}

func (g *Game) NextItem() {
	currentRound, err := g.GetCurrentRound()
	if err != nil {
		return
	}

	var totalItems int
	switch currentRound.Type {
	case RoundTypeMultipleChoice, RoundTypeFastestBuzz:
		totalItems = len(currentRound.Questions)
	case RoundTypeChallenge:
		totalItems = len(currentRound.Challenges)
	default:
		return
	}
	if totalItems == 0 {
		return
	}
	g.AnswerPending = false
	g.BuzzedPlayerID = nil
	if g.CurrentItem < totalItems-1 {
		g.CurrentItem++
		return
	}

	if g.CurrentRound < len(g.Rounds) {
		g.CurrentRound++
		g.CurrentItem = 0
		return
	}
	g.State = GameStateFinished
}

func (g *Game) Buzz(playerID uint) error {
	if g.State != GameStateInProgress {
		return ErrGameNotInProgress
	}
	currentRound, err := g.GetCurrentRound()
	if err != nil {
		return ErrNoCurrentRound
	}
	if currentRound.Type != RoundTypeFastestBuzz {
		return ErrWrongRoundType
	}
	var player *uint
	for i := range g.Players {
		if g.Players[i].PlayerID == playerID {
			player = &g.Players[i].PlayerID
			break
		}
	}
	if player == nil {
		return ErrPlayerNotFound
	}

	if g.BuzzedPlayerID != nil {
		return ErrAlreadyBuzzed
	}

	g.BuzzedPlayerID = player

	return nil
}

func (g *Game) SubmitAnswer(playerID uint, answer string) (bool, error) {
	if g.State != GameStateInProgress {
		return false, ErrGameNotInProgress
	}

	if g.AnswerPending {
		return false, ErrAnswerPending
	}
	currentRound, err := g.GetCurrentRound()
	if err != nil {
		return false, ErrNoCurrentRound
	}
	var question Question
	if currentRound.Type == RoundTypeMultipleChoice || currentRound.Type == RoundTypeFastestBuzz {
		question, err = g.GetCurrentQuestion()
		if err != nil {
			return false, ErrNoCurrentQuestion
		}

	}

	var playerTeamID *uint
	for i := range g.Players {
		if g.Players[i].PlayerID == playerID {
			playerTeamID = g.Players[i].TeamID
			break
		}
	}
	if playerTeamID == nil {
		return false, ErrPlayerNotFound
	}

	switch currentRound.Type {
	case RoundTypeMultipleChoice:
		correct := question.CorrectAnswer == answer
		if correct {
			for j := range g.Teams {
				if g.Teams[j].TeamID == *playerTeamID {
					g.Teams[j].Points += 100
				}
			}
		}
		g.AnswerPending = true
		return correct, nil
	case RoundTypeFastestBuzz:
		if g.BuzzedPlayerID == nil {
			return false, ErrNotBuzzedPlayer
		}
		if *g.BuzzedPlayerID != playerID {
			return false, ErrNotBuzzedPlayer
		}

		for i := range g.Teams {
			if g.Teams[i].TeamID == *playerTeamID {
				if question.CorrectAnswer != answer {
					g.Teams[i].Points -= 100
				} else {
					g.Teams[i].Points += 100
				}
				break
			}
		}
		g.AnswerPending = true
		return question.CorrectAnswer == answer, nil
	case RoundTypeChallenge:
		return false, ErrUnsupportedRoundType
	default:
		return false, ErrInvalidRoundType
	}
}

func (g *Game) CompleteChallenge(teamPoints map[uint]int) error {
	if g.State != GameStateInProgress {
		return ErrGameNotInProgress
	}
	round, err := g.GetCurrentRound()
	if err != nil {
		return ErrNoCurrentRound
	}
	if round.Type != RoundTypeChallenge {
		return ErrWrongRoundType
	}

	_, err = g.GetCurrentChallenge()
	if err != nil {
		return ErrNoCurrentChallenge
	}

	for teamID := range teamPoints {
		found := false

		for i := range g.Teams {
			if g.Teams[i].TeamID == teamID {
				found = true
				break
			}
		}
		if !found {
			return ErrInvalidTeamPoints
		}
	}

	for i := range g.Teams {
		teamID := g.Teams[i].TeamID
		_, exists := teamPoints[teamID]
		if !exists {
			return ErrInvalidTeamPoints
		}
	}

	for i := range g.Teams {
		teamID := g.Teams[i].TeamID
		pointsToAward := teamPoints[teamID]
		g.Teams[i].Points += pointsToAward
	}
	return nil
}

func (g *Game) StartGame() error {
	if len(g.Players) < 2 {
		g.State = GameStateLobby
		return ErrNotEnoughPlayers
	}

	g.createTeams()

	g.assignPlayersToTeams()

	g.CurrentRound = 1
	g.CurrentItem = 0
	g.State = GameStateInProgress

	return nil
}
