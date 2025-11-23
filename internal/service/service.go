package service

import (
	"github.com/vnchk1/pr-manager/internal/repository"
)

type Service struct {
	User  UserService
	Team  TeamService
	PR    PRService
	Stats StatsService
}

func New(repo *repository.Repository) *Service {
	reviewerSelector := NewReviewerSelector(repo.User)

	return &Service{
		User:  NewUserService(repo.User),
		Team:  NewTeamService(repo.Team, repo.User),
		PR:    NewPRService(repo.PullRequest, repo.User, repo.Team, reviewerSelector),
		Stats: NewStatsService(repo.Stats, repo.User),
	}
}
