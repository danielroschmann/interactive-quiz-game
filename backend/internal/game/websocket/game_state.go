package websocket

import (
	"encoding/json"

	"github.com/danielroschmann/interactive-quiz-game/backend/internal/game"
)

func BuildGameStatePayload(g *game.Game) (GameStatePayload, error) {
	payload := GameStatePayload{
		State:        string(g.State),
		CurrentRound: g.CurrentRound,
		CurrentItem:  g.CurrentItem,
		Players:      make([]PlayerView, 0, len(g.Players)),
		Teams:        make([]TeamView, 0, len(g.Teams)),
	}

	for i := range g.Players {
		payload.Players = append(payload.Players, PlayerView{
			PlayerID:   g.Players[i].PlayerID,
			PlayerName: g.Players[i].PlayerName,
			TeamID:     g.Players[i].TeamID,
		})
	}
	for i := range g.Teams {
		payload.Teams = append(payload.Teams, TeamView{
			TeamID:   g.Teams[i].TeamID,
			TeamName: g.Teams[i].TeamName,
			Points:   g.Teams[i].Points,
		})
	}

	if g.State != game.GameStateInProgress {
		return payload, nil
	}

	round, err := g.GetCurrentRound()
	if err != nil {
		return GameStatePayload{}, game.ErrNoCurrentRound
	}

	payload.RoundType = string(round.Type)

	switch round.Type {
	case game.RoundTypeMultipleChoice, game.RoundTypeFastestBuzz:
		question, err := g.GetCurrentQuestion()
		if err != nil {
			return GameStatePayload{}, game.ErrNoCurrentQuestion
		}
		if g.BuzzedPlayerID != nil {
			playerID := *g.BuzzedPlayerID
			payload.BuzzedPlayerID = &playerID
		}
		payload.Question = &QuestionView{
			Question:      question.Question,
			AnswerOptions: append([]string(nil), question.AnswerOptions...),
		}
	case game.RoundTypeChallenge:
		challenge, err := g.GetCurrentChallenge()
		if err != nil {
			return GameStatePayload{}, game.ErrNoCurrentChallenge
		}
		payload.Challenge = &ChallengeView{
			Title:       challenge.Title,
			Description: challenge.Description,
		}
	default:
		return GameStatePayload{}, game.ErrWrongRoundType
	}
	return payload, nil
}

func BuildGameStateUpdateMessage(g *game.Game) ([]byte, error) {
	payload, err := BuildGameStatePayload(g)
	if err != nil {
		return nil, err
	}

	payloadData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Type:    GameStateUpdate,
		Payload: payloadData,
	}

	msgData, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return msgData, nil
}

func (h *Hub) BroadcastGameState(g *game.Game) error {
	h.mu.Lock()
	msg, err := BuildGameStateUpdateMessage(g)
	h.mu.Unlock()
	if err != nil {
		return err
	}

	h.broadcast <- msg
	return nil
}
