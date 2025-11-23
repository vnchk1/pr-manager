package service

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"github.com/vnchk1/pr-manager/internal/repository"
	"context"
)

type TeamService interface {
	Create(ctx context.Context, team *models.Team) (*models.Team, error)
	Get(ctx context.Context, teamName string) (*models.Team, error)
}

type teamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository) TeamService {
	return &teamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *teamService) Create(ctx context.Context, team *models.Team) (*models.Team, error) {
	if err := team.Validate(); err != nil {
		return nil, err
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	return s.teamRepo.GetByName(ctx, team.Name)
}

func (s *teamService) Get(ctx context.Context, teamName string) (*models.Team, error) {
	return s.teamRepo.GetByName(ctx, teamName)
}
