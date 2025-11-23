package service

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"github.com/vnchk1/pr-manager/internal/repository"
	"context"
)

type UserService interface {
	SetActive(ctx context.Context, userID string, isActive bool) (*models.User, error)
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetReviewPRs(ctx context.Context, userID string) ([]*models.PullRequestShort, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) SetActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err = s.userRepo.SetActive(ctx, userID, isActive); err != nil {
		return nil, err
	}

	updatedUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *userService) GetByID(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *userService) GetReviewPRs(ctx context.Context, userID string) ([]*models.PullRequestShort, error) {

	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	return []*models.PullRequestShort{}, nil
}
