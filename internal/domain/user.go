package domain

import "pr-manage-service/internal/interfaces/dto"

type User struct {
	UserID   string
	UserName string
	TeamName string
	IsActive bool
}

type UserService interface {
	SetIsActive(teamName, userID string, v bool) (string, error)
	GetReview(teamName, userID string) (*dto.UserPRsResponse, error)
}

type UserRepository interface {
	ChangeActive(teamName, userID string, isActive bool) (name string, err error)
}
