package dto

type TeamRequest struct {
	TeamName string   `json:"team_name"`
	Members  []Member `json:"members"`
}

type Member struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResponse struct {
	Team TeamRequest `json:"team"`
}
