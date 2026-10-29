package websocket

import "encoding/json"

type MessageType string

const (
	JoinGame          MessageType = "join_game"
	LeaveGame         MessageType = "leave_game"
	Buzz              MessageType = "buzz"
	Answer            MessageType = "answer"
	PlayerJoined      MessageType = "player_joined"
	PlayerLeft        MessageType = "player_left"
	GameStateUpdate   MessageType = "game_state_update"
	StartGame         MessageType = "start_game"
	CompleteChallenge MessageType = "complete_challenge"
	AnswerResult      MessageType = "answer_result"
	MakeTeams         MessageType = "make_teams"
)

type PlayerView struct {
	PlayerID   uint   `json:"player_id"`
	PlayerName string `json:"player_name"`
	TeamID     *uint  `json:"team_id"`
}

type TeamView struct {
	TeamID   uint   `json:"team_id"`
	TeamName string `json:"team_name"`
	Points   int    `json:"points"`
}

type QuestionView struct {
	Question      string   `json:"question"`
	AnswerOptions []string `json:"answer_options,omitempty"`
}

type ChallengeView struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type AnswerPayload struct {
	Answer string `json:"answer"`
}

type AnswerResultPayload struct {
	PlayerID      uint   `json:"player_id"`
	IsCorrect     bool   `json:"is_correct"`
	CorrectAnswer string `json:"correct_answer"`
}

type CompleteChallengePayload struct {
	TeamPoints map[uint]int `json:"team_points"`
}

type GameStatePayload struct {
	State          string         `json:"state"`
	Players        []PlayerView   `json:"players"`
	Teams          []TeamView     `json:"teams"`
	CurrentRound   int            `json:"current_round"`
	CurrentItem    int            `json:"current_item"`
	RoundType      string         `json:"round_type"`
	Question       *QuestionView  `json:"question,omitempty"`
	Challenge      *ChallengeView `json:"challenge,omitempty"`
	BuzzedPlayerID *uint          `json:"buzzed_player_id,omitempty"`
	HostPlayerID   uint           `json:"host_player_id"`
}

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
