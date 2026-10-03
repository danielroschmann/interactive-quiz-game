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
}
