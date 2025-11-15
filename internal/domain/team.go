package domain

import (
	"pr-manage-service/internal/interfaces/dto"
)

type Team struct {
	TeamName string
	Members  []User
}

type TeamService interface {
	AddTeam(team *dto.TeamRequest) error
	GetTeamByName(teamName string) (*dto.TeamResponse, error)
}

type TeamRepository interface {
	AddNewTeam(teamName string, members *[]User) error
	GetTeamInfoByName(teamName string) (*Team, error)
}
