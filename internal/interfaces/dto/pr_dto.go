package dto

import (
	"time"
)

type PRCreateRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	TeamName        string `json:"team_name"`
}

type PRResponse struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	TeamName          string   `json:"team_name"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers,omitempty"`
}

type PRMergeResponse struct {
	*PRResponse
	MergedAt time.Time `json:"mergedAt"`
}

type PRReassignResponse struct {
	PR         PRResponse `json:"pr"`
	ReplacedBy string     `json:"replaced_by"`
}
