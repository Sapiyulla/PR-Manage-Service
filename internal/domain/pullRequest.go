package domain

import (
	"pr-manage-service/internal/interfaces/dto"
	"time"
)

type STATUS string

const (
	OPEN   STATUS = "OPEN"
	MERGED STATUS = "MERGED"
)

type PullRequest struct {
	PrID              string
	PrName            string
	AuthorID          string
	TeamName          string
	Status            STATUS
	NeedMoreReviewers bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type PRService interface {
	GetPRsByUser(userID, teamName string) (*dto.UserPRsResponse, error)
	Create(*dto.PRCreateRequest) (*dto.PRResponse, error)
	Merge(*dto.PRCreateRequest) (*dto.PRMergeResponse, error)
	Reassign(prID string, oldRevID string) (*dto.PRReassignResponse, error)
}

type PRRepository interface {
	GetWithUser(*User) (*[]PullRequest, error)
	CreateNewPR(*PullRequest) (assigned_reviewers []string, err error)
	Merge(prID string) (pr *PullRequest, assigned_reviewers []string, err error)
	Reassign(prID string, userID string) (pr *PullRequest, assigned_reviewers []string, replacedUserID string, err error)
}
