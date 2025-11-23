package repository

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetByTeam(ctx context.Context, teamName string) ([]*models.User, error)
	Update(ctx context.Context, user *models.User) error
	SetActive(ctx context.Context, userID string, isActive bool) error
	GetActiveTeamMembers(ctx context.Context, teamName string) ([]*models.User, error)
	GetActiveTeamMembersExcluding(ctx context.Context, teamName string, excludeUserIDs []string) ([]*models.User, error)
}

type TeamRepository interface {
	Create(ctx context.Context, team *models.Team) error
	GetByName(ctx context.Context, teamName string) (*models.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
	Update(ctx context.Context, team *models.Team) error
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *models.PullRequest) error
	GetByID(ctx context.Context, prID string) (*models.PullRequest, error)
	GetByAuthor(ctx context.Context, authorID string) ([]*models.PullRequest, error)
	GetByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequestShort, error)
	Update(ctx context.Context, pr *models.PullRequest) error
	Merge(ctx context.Context, prID string, mergedAt time.Time) error
	Exists(ctx context.Context, prID string) (bool, error)
	GetOpenPRsWithReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequest, error)
}

type StatsRepository interface {
	GetAssignmentStats(ctx context.Context) ([]*models.UserAssignmentStats, error)
	GetPRAssignmentStats(ctx context.Context) (*models.PRAssignmentStats, error)
	GetUserAssignmentCount(ctx context.Context, userID string) (int, error)
}

type Repository struct {
	User        UserRepository
	Team        TeamRepository
	PullRequest PullRequestRepository
	Stats       StatsRepository
}
