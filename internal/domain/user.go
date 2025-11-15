package domain

import "pr-manage-service/internal/interfaces/dto"

type User struct {
	UserID   string
	UserName string
	IsActive bool
}

type UserService interface {
	SetIsActive(teamName, userID string, v bool) (string, error)
	GetReview(teamName, userID string) (*dto.PullRequestShortResponse, error)
}

type UserRepository interface {
	AddNew(user *User) error
	ChangeActive(teamName, userID string, isActive bool) error
	GetByTeamAndID(teamName, userID string) (*User, error)
	GetManyByTeamAndID(teamName, userID string) (*[]User, error)
}
