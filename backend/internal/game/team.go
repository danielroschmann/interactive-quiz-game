package game

type Team struct {
	TeamID    uint   `json:"team_id"`
	TeamName  string `json:"team_name"`
	PlayerIDs []uint `json:"player_ids"`
	Points    int    `json:"points"`
}
