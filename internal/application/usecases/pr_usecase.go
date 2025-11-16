package usecases

import (
	"pr-manage-service/internal/domain"
	"pr-manage-service/internal/interfaces/dto"
)

type prUseCase struct {
	repo domain.PRRepository
}

// GetPRsByUser implements domain.PRService.
func (p *prUseCase) GetPRsByUser(userID string, teamName string) (*dto.UserPRsResponse, error) {
	if PRs, err := p.repo.GetWithUser(&domain.User{UserID: userID, TeamName: teamName}); err != nil {
		return nil, err
	} else {
		prResp := make([]dto.PRResponse, 0, len(*PRs))
		for _, pr := range *PRs {
			prResp = append(prResp, dto.PRResponse{
				PullRequestID:   pr.PrID,
				PullRequestName: pr.PrName,
				AuthorID:        pr.AuthorID,
				TeamName:        pr.TeamName,
				Status:          string(pr.Status),
			})
		}
		return &dto.UserPRsResponse{
			UserID:       userID,
			PullRequests: prResp,
		}, nil
	}
}

func NewPrUseCase(repo domain.PRRepository) domain.PRService {
	return &prUseCase{
		repo: repo,
	}
}

// Create implements domain.PRService.
func (p *prUseCase) Create(req *dto.PRCreateRequest) (*dto.PRResponse, error) {
	pr := &domain.PullRequest{
		PrID:     req.PullRequestID,
		PrName:   req.PullRequestName,
		AuthorID: req.AuthorID,
		TeamName: req.TeamName,
	}
	if assigned_revs, err := p.repo.CreateNewPR(pr); err != nil {
		return nil, err
	} else {
		return &dto.PRResponse{
			PullRequestID:     req.PullRequestID,
			PullRequestName:   req.PullRequestName,
			AuthorID:          req.AuthorID,
			TeamName:          req.TeamName,
			Status:            string(pr.Status),
			AssignedReviewers: assigned_revs,
		}, nil
	}
}

// Merge implements domain.PRService.
func (p *prUseCase) Merge(req *dto.PRCreateRequest) (*dto.PRMergeResponse, error) {
	if pr, ar, err := p.repo.Merge(req.PullRequestID); err != nil {
		return nil, err
	} else {
		return &dto.PRMergeResponse{
			PRResponse: &dto.PRResponse{
				PullRequestID:     req.PullRequestID,
				PullRequestName:   pr.PrName,
				TeamName:          pr.TeamName,
				Status:            string(pr.Status),
				AssignedReviewers: ar,
			},
			MergedAt: pr.UpdatedAt,
		}, nil
	}
}

// Reassign implements domain.PRService.
func (p *prUseCase) Reassign(prID string, oldRevID string) (*dto.PRReassignResponse, error) {
	panic("unimplemented")
}
