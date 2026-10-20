package websocket

import (
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
