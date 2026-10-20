package websocket

import (
	"bytes"
	"encoding/json"
	"slices"
	"testing"

	"github.com/danielroschmann/interactive-quiz-game/backend/internal/game"
)

func initializeGame(t *testing.T) *game.Game {
	t.Helper()

	rounds, err := game.LoadRounds("../questions.json")
	if err != nil {
		t.Fatalf("expected to load rounds, got %v", err)
	}

	g := game.NewGame(rounds)

	g.JoinGame("John")
	g.JoinGame("Doe")
	g.JoinGame("Foo")
	g.JoinGame("Bar")
	g.JoinGame("Troy")

	if err := g.StartGame(); err != nil {
		t.Fatalf("expected game to start, got %v", err)
	}

	return g
}

func TestBuildGameStatePayload_MultipleChoiceRound(t *testing.T) {
	game := initializeGame(t)

	game.CurrentRound = 1
	game.CurrentItem = 0

	payload, err := BuildGameStatePayload(game)
	if err != nil {
		t.Fatalf("could not build game state payload %v", err)
	}

	if payload.State != "in_progress" {
		t.Errorf("expected state in_progress, got %s", payload.State)
	}

	if payload.CurrentRound != 1 {
		t.Errorf("expected current round 1, got %d", payload.CurrentRound)
	}

	if payload.CurrentItem != 0 {
		t.Errorf("expected current item 0, got %d", payload.CurrentItem)
	}

	if payload.RoundType != "multiple_choice" {
		t.Errorf("expected round type multiple_choice, got %s", payload.RoundType)
	}

	playersPayload := len(payload.Players)
	teamsPayload := len(payload.Teams)

	if playersPayload != len(game.Players) {
		t.Errorf("expected same amount of players, got %d and %d", playersPayload, len(game.Players))
	}

	if teamsPayload != len(game.Teams) {
		t.Errorf("expected same amount of teams, got %d and %d", teamsPayload, len(game.Teams))
	}

	if payload.Question == nil {
		t.Fatalf("expected a question in payload")
	}

	if payload.Challenge != nil {
		t.Errorf("expected challenge to be nil")
	}

	question, err := game.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get question %v", err)
	}

	if payload.Question.Question != question.Question {
		t.Errorf("expected question to be the same, got %s and %s", payload.Question.Question, question.Question)
	}

	if !slices.Equal(payload.Question.AnswerOptions, question.AnswerOptions) {
		t.Errorf("expected answer options to be the same")
	}

	p, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("could not marshal payload %v", err)
	}

	if bytes.Contains(p, []byte(`"correct_answer"`)) {
		t.Errorf("payload must not expose the correct answer")
	}
}
