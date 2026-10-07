package game

import (
	"errors"
	"testing"
)

func TestStartGame_WithTwoPlayers_StartsGame(t *testing.T) {
	rounds, err := LoadRounds("questions.json")
	if err != nil {
		t.Fatalf("expected rounds to load, got err %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")

	err = game.StartGame()
	if err != nil {
		t.Errorf("expected game to start, got error: %v", err)
	}

	if game.CurrentRound != 1 {
		t.Errorf("expected current round to be %d got %d", 1, game.CurrentRound)
	}

	if game.State != GameStateInProgress {
		t.Errorf("expected game state to be in progress, got %s", game.State)
	}
}

func TestStartGame_WithOnePlayer_ReturnsError(t *testing.T) {
	rounds, err := LoadRounds("questions.json")
	if err != nil {
		t.Fatalf("expected rounds to load, got err %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")

	err = game.StartGame()
	if !errors.Is(err, ErrNotEnoughPlayers) {
		t.Errorf("expected ErrNotEnoughPlayers, got %v", err)
	}

	if game.State != GameStateLobby {
		t.Errorf("expected game state to remain in lobby, got %s", game.State)
	}
}

func TestStartGame_WithFourPlayers_CreatesTwoTeams(t *testing.T) {
	rounds, err := LoadRounds("questions.json")
	if err != nil {
		t.Fatalf("expected rounds to load, got err %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")

	err = game.StartGame()
	if err != nil {
		t.Fatalf("expected game to start, got err %v", err)
	}

	if len(game.Teams) != 2 {
		t.Errorf("expected 2 teams, got %d", len(game.Teams))
	}
}

func TestStartGame_WithFourPlayers_DistributesAllPlayers(t *testing.T) {
	rounds, err := LoadRounds("questions.json")
	if err != nil {
		t.Fatalf("expected rounds to load, got err %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")

	err = game.StartGame()
	if err != nil {
		t.Fatalf("expected game to start, got err %v", err)
	}

	totalPlayers := 0
	for _, team := range game.Teams {
		totalPlayers += len(team.PlayerIDs)
	}

	if totalPlayers != len(game.Players) {
		t.Errorf("expected %d players, got %d players", len(game.Players), totalPlayers)
	}

	playerMap := make(map[uint]bool)

	for _, team := range game.Teams {
		for _, playerID := range team.PlayerIDs {
			if playerMap[playerID] {
				t.Errorf("player %d was assigned to multiple teams", playerID)
			}
			playerMap[playerID] = true
		}
	}

	for _, player := range game.Players {
		if !playerMap[player.PlayerID] {
			t.Errorf("player %d was not assigned to a team", player.PlayerID)
		}
	}
}

func TestStartGame_WithFourPlayers_DistributesPlayersEvenly(t *testing.T) {
	rounds, err := LoadRounds("questions.json")
	if err != nil {
		t.Fatalf("expected rounds to load, got err %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")

	err = game.StartGame()
	if err != nil {
		t.Errorf("expected game to start, got err %v", err)
	}

	if len(game.Teams[0].PlayerIDs) != 2 || len(game.Teams[1].PlayerIDs) != 2 {
		t.Errorf("expected two members on each team, got %d on the first and %d on the second", len(game.Teams[0].PlayerIDs), len(game.Teams[1].PlayerIDs))
	}
}

func TestStartGame_WithFivePlayers_DistributesPlayersCorrect(t *testing.T) {
	rounds, err := LoadRounds("questions.json")
	if err != nil {
		t.Fatalf("expected rounds to load, got err %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")
	game.JoinGame("Troy")

	err = game.StartGame()
	if err != nil {
		t.Errorf("expected game to start, got err %v", err)
	}

	firstTeamLength := len(game.Teams[0].PlayerIDs)
	secondTeamLength := len(game.Teams[1].PlayerIDs)
	if !((firstTeamLength == 3 && secondTeamLength == 2) ||
		(firstTeamLength == 2 && secondTeamLength == 3)) {
		t.Errorf(
			"expected teams to have 3 and 2 players, got %d and %d",
			firstTeamLength,
			secondTeamLength,
		)
	}
}

func TestStartGame_AssignsTeamIDToEveryPlayer(t *testing.T) {
	rounds, err := LoadRounds("questions.json")
	if err != nil {
		t.Fatalf("expected rounds to load, got err %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")
	game.JoinGame("Troy")

	err = game.StartGame()
	if err != nil {
		t.Errorf("expected game to start, got err %v", err)
	}

	for _, p := range game.Players {
		if p.TeamID == nil {
			t.Errorf("expected player %d to have team id", p.PlayerID)
		}
	}
}

func TestLoadRounds_WithValidJSON(t *testing.T) {
	fileName := "questions.json"
	rounds, err := LoadRounds(fileName)
	if err != nil {
		t.Fatalf("expected to load rounds, got %v", err)
	}

	if len(rounds) != 2 {
		t.Errorf("expected 2 rounds, got %d", len(rounds))
	}

	if rounds[0].Type != RoundTypeMultipleChoice {
		t.Errorf("expected type %s got %s", RoundTypeMultipleChoice, rounds[0].Type)
	}

	if len(rounds[0].Questions) != 2 {
		t.Errorf("expected 2 questions, got %d", len(rounds[0].Questions))
	}

	if rounds[0].Questions[1].CorrectAnswer != "Baida" {
		t.Errorf("expected correct answer %s got %s", "Baida", rounds[0].Questions[1].CorrectAnswer)
	}
}

func TestStartGame_GameStartsAtFirstQuestion(t *testing.T) {
	fileName := "questions.json"
	rounds, err := LoadRounds(fileName)
	if err != nil {
		t.Fatalf("expected to load rounds, got %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")
	game.JoinGame("Troy")

	err = game.StartGame()
	if err != nil {
		t.Errorf("expected game to start, got err %v", err)
	}

	if game.CurrentQuestion != 0 {
		t.Errorf("expected current question to be index 0, got %d", game.CurrentQuestion)
	}
}

func TestGetCurrentQuestion_ReturnsFirstQuestion(t *testing.T) {
	fileName := "questions.json"
	rounds, err := LoadRounds(fileName)
	if err != nil {
		t.Fatalf("expected to load rounds, got %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")
	game.JoinGame("Troy")

	err = game.StartGame()
	if err != nil {
		t.Fatalf("expected game to start, got err %v", err)
	}

	expectedQuestion := rounds[0].Questions[0]
	currentQuestion, err := game.GetCurrentQuestion()
	if err != nil {
		t.Errorf("expected to get current question, got err %v", err)
	}

	if currentQuestion.Question != expectedQuestion.Question {
		t.Errorf("expected question %s, got %s", expectedQuestion.Question, currentQuestion.Question)
	}
}

func TestGetCurrentQuestion_ReturnsErrorWhenIndexOutOfRange(t *testing.T) {
	fileName := "questions.json"
	rounds, err := LoadRounds(fileName)
	if err != nil {
		t.Fatalf("expected to load rounds, got %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")
	game.JoinGame("Troy")

	err = game.StartGame()
	if err != nil {
		t.Fatalf("expected game to start, got err %v", err)
	}

	game.CurrentQuestion = 99

	_, err = game.GetCurrentQuestion()
	if !errors.Is(err, ErrNoCurrentQuestion) {
		t.Errorf("expected error %v got %v", ErrNoCurrentQuestion, err)
	}
}

func TestGetNextQuestion_GoesFromQuestionZeroToOne(t *testing.T) {
	fileName := "questions.json"
	rounds, err := LoadRounds(fileName)
	if err != nil {
		t.Fatalf("expected to load rounds, got %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")
	game.JoinGame("Troy")

	err = game.StartGame()
	if err != nil {
		t.Fatalf("expected game to start, got err %v", err)
	}

	game.GetNextQuestion()
	if game.CurrentQuestion != 1 {
		t.Errorf("expected current question to be 1, got %d", game.CurrentQuestion)
	}
}

func TestGetNextQuestion_GoesFromRoundOneToTwo(t *testing.T) {
	fileName := "questions.json"
	rounds, err := LoadRounds(fileName)
	if err != nil {
		t.Fatalf("expected to load rounds, got %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")
	game.JoinGame("Troy")

	err = game.StartGame()
	if err != nil {
		t.Fatalf("expected game to start, got err %v", err)
	}

	game.CurrentQuestion = len(rounds[0].Questions) - 1

	game.GetNextQuestion()

	if game.CurrentRound != 2 {
		t.Errorf("expected round to be 2, got %d", game.CurrentRound)
	}

	if game.CurrentQuestion != 0 {
		t.Errorf("expected current question to be 0, got %d", game.CurrentQuestion)
	}
}

func TestGetNextQuestion_ChangeStateToGameStateFinished(t *testing.T) {
	fileName := "questions.json"
	rounds, err := LoadRounds(fileName)
	if err != nil {
		t.Fatalf("expected to load rounds, got %v", err)
	}
	game := NewGame(rounds)
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")
	game.JoinGame("Troy")

	err = game.StartGame()
	if err != nil {
		t.Fatalf("expected game to start, got err %v", err)
	}

	lastRoundIndex := len(game.Rounds) - 1

	game.CurrentQuestion = len(rounds[lastRoundIndex].Questions) - 1
	game.GetNextQuestion()

	if game.State != GameStateFinished {
		t.Errorf("expected game state to be finished, got %s ", game.State)
	}
}
