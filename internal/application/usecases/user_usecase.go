package usecases

import (
	"pr-manage-service/internal/domain"
	"pr-manage-service/internal/interfaces/dto"
)

type useUseCase struct {
	repo domain.UserRepository
	prUC domain.PRService
}

func NewUserUseCase(repo domain.UserRepository, prUC domain.PRService) domain.UserService {
	return &useUseCase{
		repo: repo,
		prUC: prUC,
	}
}

// GetReview implements domain.UserService.
func (u *useUseCase) GetReview(teamName string, userID string) (*dto.UserPRsResponse, error) {
	return u.prUC.GetPRsByUser(userID, teamName)
}

// SetIsActive implements domain.UserService.
func (u *useUseCase) SetIsActive(teamName string, userID string, v bool) (string, error) {
	return u.repo.ChangeActive(teamName, userID, v)
}
