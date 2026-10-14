package game

import (
	"errors"
	"testing"
)

func initializeGame(t *testing.T) *Game {
	t.Helper()

	rounds, err := LoadRounds("questions.json")
	if err != nil {
		t.Fatalf("expected to load rounds, got %v", err)
	}

	game := NewGame(rounds)

	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")
	game.JoinGame("Troy")

	if err := game.StartGame(); err != nil {
		t.Fatalf("expected game to start, got %v", err)
	}

	return game
}

func initializeBuzzRoundGame(t *testing.T) *Game {
	t.Helper()

	game := initializeGame(t)
	game.CurrentRound = 2

	round, err := game.GetCurrentRound()
	if err != nil {
		t.Fatalf("could not get current round: %v", err)
	}
	if round.Type != RoundTypeFastestBuzz {
		t.Fatalf("expected round type %s, got %s",
			RoundTypeFastestBuzz, round.Type)
	}

	return game
}

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

	if len(rounds) != 3 {
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

	if game.CurrentItem != 0 {
		t.Errorf("expected current question to be index 0, got %d", game.CurrentItem)
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

	game.CurrentItem = 99

	_, err = game.GetCurrentQuestion()
	if !errors.Is(err, ErrNoCurrentQuestion) {
		t.Errorf("expected error %v got %v", ErrNoCurrentQuestion, err)
	}
}

func TestGetCurrentRound_ReturnsFirstRoundAndCorrectRoundType(t *testing.T) {
	game := initializeGame(t)

	expectedRound := 1
	currentRound, err := game.GetCurrentRound()
	if err != nil {
		t.Fatalf("expected to get current round, got %v", err)
	}

	if currentRound.RoundNumber != expectedRound {
		t.Errorf("expected current round number to be %d, got %d", expectedRound, currentRound.RoundNumber)
	}

	if currentRound.Type != RoundTypeMultipleChoice {
		t.Errorf("expected round type to be %s, got round type %s", RoundTypeMultipleChoice, currentRound.Type)
	}
}

func TestGetCurrentRound_ReturnsErrorWhenIndexOutOfRange(t *testing.T) {
	game := initializeGame(t)

	game.CurrentRound = 99

	_, err := game.GetCurrentRound()
	if !errors.Is(err, ErrNoCurrentRound) {
		t.Errorf("expected error %v got %v", ErrNoCurrentRound, err)
	}
}

func TestNextQuestion_GoesFromQuestionZeroToOne(t *testing.T) {
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

	game.NextItem()
	if game.CurrentItem != 1 {
		t.Errorf("expected current question to be 1, got %d", game.CurrentItem)
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

	game.CurrentItem = len(rounds[0].Questions) - 1

	game.NextItem()

	if game.CurrentRound != 2 {
		t.Errorf("expected round to be 2, got %d", game.CurrentRound)
	}

	if game.CurrentItem != 0 {
		t.Errorf("expected current question to be 0, got %d", game.CurrentItem)
	}
}

func TestNextQuestion_ChangeStateToGameStateFinished(t *testing.T) {
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
	game.CurrentRound = lastRoundIndex + 1
	game.CurrentItem = len(rounds[lastRoundIndex].Challenges) - 1
	game.NextItem()

	if game.State != GameStateFinished {
		t.Errorf("expected game state to be finished, got %s ", game.State)
	}
}

func TestSubmitAnswer_AssignPointsForCorrectAnswer(t *testing.T) {
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

	_, err = game.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get current question")
	}
	answer, err := game.SubmitAnswer(1, "10")
	if err != nil {
		t.Errorf("expected to get current question, got %v", err)
	}

	if !answer {
		t.Errorf("expected answer to be correct")
	}

	playerID := uint(1)
	for i := range game.Players {
		if game.Players[i].PlayerID == playerID {
			playerTeamID := game.Players[i].TeamID

			for j := range game.Teams {
				if game.Teams[j].TeamID == *playerTeamID {
					if game.Teams[j].Points != 100 {
						t.Errorf("expected team to have 100 points, got %d", game.Teams[j].Points)
					}
				}
			}
		}
	}
}

func TestSubmitAnswer_ReturnsFalseAndZeroPointsForWrongAnswer(t *testing.T) {
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

	_, err = game.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get current question")
	}
	answer, err := game.SubmitAnswer(1, "13")
	if err != nil {
		t.Errorf("expected to get current question, got %v", err)
	}

	if answer {
		t.Errorf("expected answer to be wrong")
	}

	playerID := uint(1)
	for i := range game.Players {
		if game.Players[i].PlayerID == playerID {
			playerTeamID := game.Players[i].TeamID

			for j := range game.Teams {
				if game.Teams[j].TeamID == *playerTeamID {
					if game.Teams[j].Points != 0 {
						t.Errorf("expected team to have 0 points, got %d", game.Teams[j].Points)
					}
				}
			}
		}
	}
}

func TestSubmitAnswer_PlayerNotFoundReturnsErr(t *testing.T) {
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

	_, err = game.SubmitAnswer(999, "10")

	if !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("expected error %v got error %v", ErrPlayerNotFound, err)
	}
}

func TestBuzz_PlayerNotFound(t *testing.T) {
	game := initializeBuzzRoundGame(t)

	err := game.Buzz(999)

	if !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("expected error %v, got %v", ErrPlayerNotFound, err)
	}
}

func TestBuzz_PlayerAlreadyBuzzed(t *testing.T) {
	game := initializeBuzzRoundGame(t)

	firstPlayerID := uint(1)
	game.BuzzedPlayerID = &firstPlayerID

	err := game.Buzz(2)

	if !errors.Is(err, ErrAlreadyBuzzed) {
		t.Errorf("expected error %v, got %v", ErrAlreadyBuzzed, err)
	}
}

func TestBuzz_ReturnsErrAlreadyBuzzed(t *testing.T) {
	game := initializeBuzzRoundGame(t)

	firstPlayerID := uint(1)
	secondPlayerID := uint(2)
	err := game.Buzz(firstPlayerID)
	if err != nil {
		t.Fatalf("expected first player to buzz succesfully, got err %v", err)
	}

	err = game.Buzz(secondPlayerID)

	if !errors.Is(err, ErrAlreadyBuzzed) {
		t.Errorf("expected error %v, got %v", ErrAlreadyBuzzed, err)
	}
}

func TestBuzz_CorrectBuzz(t *testing.T) {
	game := initializeBuzzRoundGame(t)
	firstPlayerID := uint(1)
	err := game.Buzz(firstPlayerID)
	if err != nil {
		t.Fatalf("expected first player to buzz succesfully, got err %v", err)
	}

	if game.BuzzedPlayerID == nil {
		t.Errorf("expected BuzzedPlayerID to be set, got nil")
	}

	if *game.BuzzedPlayerID != firstPlayerID {
		t.Errorf("expected BuzzPLayerID to be %d, got %d", firstPlayerID, *game.BuzzedPlayerID)
	}
}

func TestSubmitAnswer_NonBuzzedPlayerReturnsError(t *testing.T) {
	game := initializeGame(t)
	game.CurrentRound = 2
	round, err := game.GetCurrentRound()
	if err != nil {
		t.Fatalf("could not get current round")
	}
	if round.Type != RoundTypeFastestBuzz {
		t.Fatalf("expected roundtype to be %s got %s", RoundTypeFastestBuzz, round.Type)
	}
	question, err := game.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get current question")
	}

	firstPlayerID := uint(1)
	secondPlayerID := uint(2)
	err = game.Buzz(firstPlayerID)
	if err != nil {
		t.Fatalf("expected first player to buzz succesfully, got err %v", err)
	}

	_, err = game.SubmitAnswer(secondPlayerID, question.CorrectAnswer)
	if !errors.Is(err, ErrNotBuzzedPlayer) {
		t.Errorf("expected err %v, got %v", ErrNotBuzzedPlayer, err)
	}
}

func TestSubmitAnswer_CorrectAnswerInBuzzRound(t *testing.T) {
	game := initializeGame(t)
	game.CurrentRound = 2
	playerID := uint(1)
	var playerTeamID *uint

	for i := range game.Players {
		if game.Players[i].PlayerID == playerID {
			playerTeamID = game.Players[i].TeamID
			break
		}
	}
	if playerTeamID == nil {
		t.Fatalf("player team ID was not found")
	}

	pointsBefore := 0
	for j := range game.Teams {
		if game.Teams[j].TeamID == *playerTeamID {
			pointsBefore = game.Teams[j].Points
		}
	}

	question, err := game.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get current question")
	}
	err = game.Buzz(playerID)
	if err != nil {
		t.Fatalf("could not buzz")
	}
	answer, err := game.SubmitAnswer(playerID, question.CorrectAnswer)
	if err != nil {
		t.Fatalf("expected to submit answer, got err %v", err)
	}

	if !answer {
		t.Fatalf("expected correct answer to return true")
	}
	for j := range game.Teams {
		if game.Teams[j].TeamID == *playerTeamID {
			if game.Teams[j].Points != pointsBefore+100 {
				t.Errorf("expected team to have %d points, got %d", pointsBefore+100, game.Teams[j].Points)
			}
		}
	}
}

func TestSubmitAnswer_WrongAnswerInBuzzRound(t *testing.T) {
	game := initializeGame(t)
	game.CurrentRound = 2
	playerID := uint(1)
	var playerTeamID *uint

	for i := range game.Players {
		if game.Players[i].PlayerID == playerID {
			playerTeamID = game.Players[i].TeamID
			break
		}
	}
	if playerTeamID == nil {
		t.Fatalf("player team ID was not found")
	}

	pointsBefore := 0
	for j := range game.Teams {
		if game.Teams[j].TeamID == *playerTeamID {
			pointsBefore = game.Teams[j].Points
		}
	}

	_, err := game.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("could not get current question")
	}
	err = game.Buzz(playerID)
	if err != nil {
		t.Fatalf("could not buzz")
	}
	roundBefore := game.CurrentRound
	questionBefore := game.CurrentItem
	answer, err := game.SubmitAnswer(playerID, "Wrong Answer")
	if err != nil {
		t.Fatalf("expected to submit answer, got err %v", err)
	}

	questionsInRound := len(game.Rounds[roundBefore-1].Questions)

	if questionBefore < questionsInRound-1 {
		if game.CurrentItem != questionBefore+1 {
			t.Errorf("expected question index %d, got %d", questionBefore+1, game.CurrentItem)
		}
	} else if roundBefore < len(game.Rounds) {
		if game.CurrentRound != roundBefore+1 {
			t.Errorf("expected round %d, got %d", roundBefore+1, game.CurrentRound)
		}
		if game.CurrentItem != 0 {
			t.Errorf("expected question index 0, got %d", game.CurrentItem)
		}
	} else {
		if game.State != GameStateFinished {
			t.Errorf("expected game state %s, got %s", GameStateFinished, game.State)
		}
	}

	if answer {
		t.Fatalf("expected wrong answer to be false")
	}
	for j := range game.Teams {
		if game.Teams[j].TeamID == *playerTeamID {
			if game.Teams[j].Points != pointsBefore-100 {
				t.Errorf("expected team to have %d points, got %d", pointsBefore-100, game.Teams[j].Points)
			}
		}
	}
}

func TestBuzz_GameNotInProgressInLobby(t *testing.T) {
	game := initializeBuzzRoundGame(t)

	game.State = GameStateLobby

	err := game.Buzz(1)

	if !errors.Is(err, ErrGameNotInProgress) {
		t.Errorf("expected error %v, got %v", ErrGameNotInProgress, err)
	}
}

func TestBuzz_GameFinished(t *testing.T) {
	game := initializeBuzzRoundGame(t)

	game.State = GameStateFinished

	err := game.Buzz(1)

	if !errors.Is(err, ErrGameNotInProgress) {
		t.Errorf("expected error %v, got %v", ErrGameNotInProgress, err)
	}
}

func TestBuzz_WrongRoundType(t *testing.T) {
	game := initializeBuzzRoundGame(t)
	// multiple choice
	game.CurrentRound = 1

	err := game.Buzz(1)

	if !errors.Is(err, ErrWrongRoundType) {
		t.Errorf("expected error %v, got %v", ErrWrongRoundType, err)
	}
}

func TestSubmitAnswer_GameNotInProgressInLobby(t *testing.T) {
	game := initializeGame(t)

	game.State = GameStateLobby

	_, err := game.SubmitAnswer(1, "random answer")

	if !errors.Is(err, ErrGameNotInProgress) {
		t.Errorf("expected error %v, got %v", ErrGameNotInProgress, err)
	}
}

func TestSubmitAnswer_GameFinished(t *testing.T) {
	game := initializeGame(t)

	game.State = GameStateFinished

	_, err := game.SubmitAnswer(1, "random answer")

	if !errors.Is(err, ErrGameNotInProgress) {
		t.Errorf("expected error %v, got %v", ErrGameNotInProgress, err)
	}
}

func TestSubmitAnswer_NoPlayerHasBuzzed(t *testing.T) {
	game := initializeBuzzRoundGame(t)

	game.BuzzedPlayerID = nil

	_, err := game.SubmitAnswer(1, "random answer")

	if !errors.Is(err, ErrNotBuzzedPlayer) {
		t.Errorf("expected error %v, got %v", ErrNotBuzzedPlayer, err)
	}
}

func TestSubmitAnswer_ChallengeRoundNotSupported(t *testing.T) {
	game := initializeGame(t)

	// challenge round
	game.CurrentRound = 3

	round, err := game.GetCurrentRound()
	if err != nil {
		t.Fatalf("could not get current round %v", err)
	}

	if round.Type != RoundTypeChallenge {
		t.Fatalf("expected round type %s, got %s", RoundTypeChallenge, round.Type)
	}
	_, err = game.SubmitAnswer(1, "random answer")

	if !errors.Is(err, ErrUnsupportedRoundType) {
		t.Errorf("expected error %v, got %v", ErrUnsupportedRoundType, err)
	}
}

func TestSubmitAnswer_InvalidRoundType(t *testing.T) {
	game := initializeGame(t)

	game.Rounds[game.CurrentRound-1].Type = RoundType("invalid_type")

	_, err := game.SubmitAnswer(1, "random answer")

	if !errors.Is(err, ErrInvalidRoundType) {
		t.Errorf("expected error %v, got %v", ErrInvalidRoundType, err)
	}
}

func TestSubmitAnswer_WrongAnswerInMultipleChoiceAdvancesToNextItem(t *testing.T) {
	game := initializeGame(t)

	currentRoundBefore := game.CurrentRound
	currentItemBefore := game.CurrentItem

	playerID := uint(1)
	var playerTeamID *uint
	for i := range game.Players {
		if game.Players[i].PlayerID == playerID {
			playerTeamID = game.Players[i].TeamID
			break
		}
	}
	if playerTeamID == nil {
		t.Fatalf("player team ID was not found")
	}

	pointsBefore := 0
	for j := range game.Teams {
		if game.Teams[j].TeamID == *playerTeamID {
			pointsBefore = game.Teams[j].Points
		}
	}

	answer, err := game.SubmitAnswer(playerID, "wrong answer")
	if err != nil {
		t.Fatalf("could not submit answer")
	}

	pointsAfter := 0
	for j := range game.Teams {
		if game.Teams[j].TeamID == *playerTeamID {
			pointsAfter = game.Teams[j].Points
			break
		}
	}

	if answer {
		t.Errorf("expected answer to be %t", answer)
	}

	if game.CurrentItem != currentItemBefore+1 {
		t.Errorf("expected current item to be %d, got %d", currentItemBefore+1, game.CurrentItem)
	}

	if game.CurrentRound != currentRoundBefore {
		t.Errorf("expected current round to stay at %d, got %d", currentRoundBefore, game.CurrentRound)
	}

	if pointsBefore != pointsAfter {
		t.Errorf("expected points to stay unchanged, got %d", pointsBefore)
	}
}
