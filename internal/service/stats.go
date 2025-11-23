package service

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"github.com/vnchk1/pr-manager/internal/repository"
	"context"
)

type StatsService interface {
	GetAssignmentStats(ctx context.Context) (*models.AssignmentStatsResponse, error)
	GetUserStats(ctx context.Context, userID string) (*models.UserAssignmentStats, error)
}

type statsService struct {
	statsRepo repository.StatsRepository
	userRepo  repository.UserRepository
}

func NewStatsService(statsRepo repository.StatsRepository, userRepo repository.UserRepository) StatsService {
	return &statsService{
		statsRepo: statsRepo,
		userRepo:  userRepo,
	}
}

func (s *statsService) GetAssignmentStats(ctx context.Context) (*models.AssignmentStatsResponse, error) {
	userStats, err := s.statsRepo.GetAssignmentStats(ctx)
	if err != nil {
		return nil, err
	}

	prStats, err := s.statsRepo.GetPRAssignmentStats(ctx)
	if err != nil {
		return nil, err
	}

	summary := s.calculateSummary(userStats, prStats)

	return &models.AssignmentStatsResponse{
		UserStats: userStats,
		PRStats:   prStats,
		Summary:   summary,
	}, nil
}

func (s *statsService) GetUserStats(ctx context.Context, userID string) (*models.UserAssignmentStats, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	assignmentCount, err := s.statsRepo.GetUserAssignmentCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.UserAssignmentStats{
		UserID:          user.ID,
		Username:        user.Username,
		TeamName:        user.TeamName,
		IsActive:        user.IsActive,
		AssignmentCount: assignmentCount,
	}, nil
}

func (s *statsService) calculateSummary(userStats []*models.UserAssignmentStats, prStats *models.PRAssignmentStats) *models.StatsSummary {
	summary := &models.StatsSummary{
		TotalUsers:       len(userStats),
		TotalAssignments: 0,
	}

	for _, stat := range userStats {
		if stat.IsActive {
			summary.ActiveUsers++
		}
		summary.TotalAssignments += stat.AssignmentCount

		if stat.AssignmentCount > summary.MostAssignments {
			summary.MostAssignments = stat.AssignmentCount
			summary.MostAssignedUser = stat.Username
		}
	}

	return summary
}
