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
