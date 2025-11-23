package service

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"github.com/vnchk1/pr-manager/internal/repository"
	"github.com/vnchk1/pr-manager/internal/utils"
	"context"

	"github.com/pkg/errors"
)

type ReviewerSelector interface {
	SelectReviewers(ctx context.Context, author *models.User) ([]string, error)
	SelectReplacementReviewer(ctx context.Context, teamName string, excludeUserIDs []string) (string, error)
}

type reviewerSelector struct {
	userRepo repository.UserRepository
}

func NewReviewerSelector(userRepo repository.UserRepository) ReviewerSelector {
	return &reviewerSelector{userRepo: userRepo}
}

func (s *reviewerSelector) SelectReviewers(ctx context.Context, author *models.User) ([]string, error) {
	// Получаем активных участников команды, исключая автора
	candidates, err := s.userRepo.GetActiveTeamMembersExcluding(
		ctx,
		author.TeamName,
		[]string{author.ID},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get team members")
	}

	if len(candidates) == 0 {
		return []string{}, nil
	}

	return s.selectRandomReviewers(candidates, 2), nil
}

func (s *reviewerSelector) SelectReplacementReviewer(
	ctx context.Context,
	teamName string,
	excludeUserIDs []string,
) (string, error) {

	candidates, err := s.userRepo.GetActiveTeamMembersExcluding(
		ctx,
		teamName,
		excludeUserIDs,
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to get team members")
	}

	if len(candidates) == 0 {
		return "", models.ErrNoCandidate
	}

	selected := s.selectRandomReviewers(candidates, 1)
	if len(selected) == 0 {
		return "", models.ErrNoCandidate
	}

	return selected[0], nil
}

func (s *reviewerSelector) selectRandomReviewers(candidates []*models.User, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	shuffled := utils.ShuffleUsers(candidates)

	count := len(shuffled)
	if count > maxCount {
		count = maxCount
	}

	reviewers := make([]string, count)
	for i := 0; i < count; i++ {
		reviewers[i] = shuffled[i].ID
	}

	return reviewers
}
