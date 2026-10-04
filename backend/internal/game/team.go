package game

type Team struct {
	TeamID    uint   `json:"team_id"`
	TeamName  string `json:"team_name"`
	PlayerIDs []uint `json:"player_ids"`
	Points    int    `json:"points"`
}

func NewTeam(id uint, name string) Team {
	return Team{
		TeamID:    id,
		TeamName:  name,
		PlayerIDs: make([]uint, 0),
		Points:    0,
	}
}
