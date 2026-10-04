package game

import (
	"errors"
	"testing"
)

func TestStartGame_WithTwoPlayers_StartsGame(t *testing.T) {
	game := NewGame()
	game.JoinGame("John")
	game.JoinGame("Doe")

	err := game.StartGame()
	if err != nil {
		t.Errorf("expected game to start, got error: %v", err)
	}

	if game.State != GameStateInProgress {
		t.Errorf("expected game state to be in progress, got %s", game.State)
	}
}

func TestStartGame_WithOnePlayer_ReturnsError(t *testing.T) {
	game := NewGame()
	game.JoinGame("John")

	err := game.StartGame()
	if !errors.Is(err, ErrNotEnoughPlayers) {
		t.Errorf("expected ErrNotEnoughPlayers, got %v", err)
	}

	if game.State != GameStateLobby {
		t.Errorf("expected game state to remain in lobby, got %s", game.State)
	}
}

func TestStartGame_WithFourPlayers_CreatesTwoTeams(t *testing.T) {
	game := NewGame()
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")

	err := game.StartGame()
	if err != nil {
		t.Fatalf("expected game to start, got err %v", err)
	}

	if len(game.Teams) != 2 {
		t.Errorf("expected 2 teams, got %d", len(game.Teams))
	}
}

func TestStartGame_WithFourPlayers_DistributesAllPlayers(t *testing.T) {
	game := NewGame()
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")

	err := game.StartGame()
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
	game := NewGame()
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")

	err := game.StartGame()
	if err != nil {
		t.Errorf("expected game to start, got err %v", err)
	}

	if len(game.Teams[0].PlayerIDs) != 2 || len(game.Teams[1].PlayerIDs) != 2 {
		t.Errorf("expected two members on each team, got %d on the first and %d on the second", len(game.Teams[0].PlayerIDs), len(game.Teams[1].PlayerIDs))
	}
}

func TestStartGame_WithFivePlayers_DistributesPlayersCorrect(t *testing.T) {
	game := NewGame()
	game.JoinGame("John")
	game.JoinGame("Doe")
	game.JoinGame("Foo")
	game.JoinGame("Bar")
	game.JoinGame("Troy")

	err := game.StartGame()
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
