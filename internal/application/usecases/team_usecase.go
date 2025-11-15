package usecases

import (
	"fmt"
	"pr-manage-service/internal/domain"
	"pr-manage-service/internal/interfaces/dto"
	"pr-manage-service/pkg/errs"
	"strings"
)

type teamUseCase struct {
	repo domain.TeamRepository
}

func NewTeamUseCase(repo domain.TeamRepository) domain.TeamService {
	return &teamUseCase{
		repo: repo,
	}
}

// AddTeam implements domain.TeamService.
func (t *teamUseCase) AddTeam(team *dto.TeamRequest) error {
	if err := validateTeam(team, false); err != nil {
		return &errs.InvalidError{
			Domain: "team",
			Desc:   err.Error(),
		}
	}
	Members := make([]domain.User, len(team.Members))
	for index, member := range team.Members {
		Members[index] = domain.User{
			UserID:   member.UserID,
			UserName: member.UserName,
			IsActive: member.IsActive,
		}
	}
	return t.repo.AddNewTeam(team.TeamName, &Members)
}

// GetTeamByName implements domain.TeamService.
func (t *teamUseCase) GetTeamByName(teamName string) (*dto.TeamResponse, error) {
	team, err := t.repo.GetTeamInfoByName(teamName)
	if err != nil {
		return nil, err
	}
	members := make([]dto.Member, len(team.Members))
	for index, user := range team.Members {
		members[index] = dto.Member{
			UserID:   user.UserID,
			UserName: user.UserName,
			IsActive: user.IsActive,
		}
	}
	return &dto.TeamResponse{
		Team: dto.TeamRequest{
			TeamName: team.TeamName,
			Members:  members,
		},
	}, nil
}

func validateTeam(team *dto.TeamRequest, onlyName bool) error {
	// Валидация названия команды
	if strings.TrimSpace(team.TeamName) == "" {
		return fmt.Errorf("team_name cannot be empty")
	}

	if len(team.TeamName) > 50 {
		return fmt.Errorf("team_name too long (max 50 characters)")
	}

	// Валидация списка участников
	if len(team.Members) == 0 {
		return fmt.Errorf("team must have at least one member")
	}

	if len(team.Members) > 100 {
		return fmt.Errorf("team cannot have more than 100 members")
	}

	if onlyName {
		return nil
	}
	// Валидация каждого участника
	userIDs := make(map[string]bool)
	for _, member := range team.Members {
		if err := validateTeamMember(member); err != nil {
			return err
		}

		// Проверка уникальности user_id в команде
		if userIDs[member.UserID] {
			return fmt.Errorf("duplicate user_id '%s' in team members", member.UserID)
		}
		userIDs[member.UserID] = true
	}

	return nil
}

func validateTeamMember(member dto.Member) error {
	// Валидация user_id
	if strings.TrimSpace(member.UserID) == "" {
		return fmt.Errorf("member:%s : user_id cannot be empty", member.UserID)
	}

	if len(member.UserID) > 50 {
		return fmt.Errorf("member:%s : user_id too long (max 50 characters)", member.UserID)
	}

	// Валидация username
	if strings.TrimSpace(member.UserName) == "" {
		return fmt.Errorf("member:%s : username cannot be empty", member.UserID)
	}

	if len(member.UserName) > 100 {
		return fmt.Errorf("member:%s : username too long (max 100 characters)", member.UserID)
	}

	// is_active всегда должен быть явно указан (т.к. required в OpenAPI)
	// В Go bool всегда имеет значение по умолчанию false, но для ясности:

	return nil
}
