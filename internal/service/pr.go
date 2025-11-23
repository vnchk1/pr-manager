package service

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"github.com/vnchk1/pr-manager/internal/repository"
	"context"
	"time"
)

type PRService interface {
	Create(ctx context.Context, req *models.PRCreateRequest) (*models.PullRequest, error)
	Merge(ctx context.Context, req *models.PRMergeRequest) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, req *models.PRReassignRequest) (*models.PullRequest, string, error)
	GetByID(ctx context.Context, prID string) (*models.PullRequest, error)
	GetByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequestShort, error)
}

type prService struct {
	prRepo           repository.PullRequestRepository
	userRepo         repository.UserRepository
	teamRepo         repository.TeamRepository
	reviewerSelector ReviewerSelector
}

func NewPRService(
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
	reviewerSelector ReviewerSelector,
) PRService {
	return &prService{
		prRepo:           prRepo,
		userRepo:         userRepo,
		teamRepo:         teamRepo,
		reviewerSelector: reviewerSelector,
	}
}

func (s *prService) Create(ctx context.Context, req *models.PRCreateRequest) (*models.PullRequest, error) {
	exists, err := s.prRepo.Exists(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, models.ErrPRExists
	}

	author, err := s.userRepo.GetByID(ctx, req.AuthorID)
	if err != nil {
		return nil, models.ErrNotFound
	}

	if !author.IsActive {
		return nil, models.ErrUserNotActive
	}

	teamExists, err := s.teamRepo.Exists(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}
	if !teamExists {
		return nil, models.ErrNotFound
	}

	reviewers, err := s.reviewerSelector.SelectReviewers(ctx, author)
	if err != nil {
		return nil, err
	}

	pr := &models.PullRequest{
		ID:                req.ID,
		Name:              req.Name,
		AuthorID:          req.AuthorID,
		Status:            models.StatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         time.Now(),
	}

	if err := pr.Validate(); err != nil {
		return nil, err
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *prService) Merge(ctx context.Context, req *models.PRMergeRequest) (*models.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if pr.Status == models.StatusMerged {
		return pr, nil
	}

	mergedAt := time.Now()
	if err := s.prRepo.Merge(ctx, req.ID, mergedAt); err != nil {
		return nil, err
	}

	mergedPR, err := s.prRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return mergedPR, nil
}

func (s *prService) ReassignReviewer(ctx context.Context, req *models.PRReassignRequest) (*models.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, "", err
	}

	if pr.Status == models.StatusMerged {
		return nil, "", models.ErrPRMerged
	}

	if !contains(pr.AssignedReviewers, req.OldReviewer) {
		return nil, "", models.ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, req.OldReviewer)
	if err != nil {
		return nil, "", models.ErrNotFound
	}

	newReviewerID, err := s.reviewerSelector.SelectReplacementReviewer(
		ctx,
		oldReviewer.TeamName,
		[]string{pr.AuthorID, req.OldReviewer},
	)
	if err != nil {
		return nil, "", err
	}

	newReviewers := make([]string, len(pr.AssignedReviewers))
	for i, reviewer := range pr.AssignedReviewers {
		if reviewer == req.OldReviewer {
			newReviewers[i] = newReviewerID
		} else {
			newReviewers[i] = reviewer
		}
	}

	pr.AssignedReviewers = newReviewers
	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, "", err
	}

	return pr, newReviewerID, nil
}

func (s *prService) GetByID(ctx context.Context, prID string) (*models.PullRequest, error) {
	return s.prRepo.GetByID(ctx, prID)
}

func (s *prService) GetByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequestShort, error) {
	if _, err := s.userRepo.GetByID(ctx, reviewerID); err != nil {
		return nil, err
	}

	return s.prRepo.GetByReviewer(ctx, reviewerID)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
