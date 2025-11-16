package dto

type UserRequest struct {
	UserID   string `json:"user_id"`
	TeamName string `json:"team_name"` // нужный параметр, без него невозможна выборка пользователей
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	*UserRequest
	UserName string `json:"username"`
}

type UserFullResponse struct {
	User UserResponse `json:"user"`
}

type UserPRsResponse struct {
	UserID       string       `json:"user_id"`
	PullRequests []PRResponse `json:"pull_requests"`
}
