package game

type Player struct {
	PlayerID   uint   `json:"player_id"`
	PlayerName string `json:"player_name"`
	TeamID     *uint  `json:"team_id,omitempty"`
}
