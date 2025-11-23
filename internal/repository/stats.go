package repository

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type statsRepository struct {
	db *pgxpool.Pool
}

func NewStatsRepository(db *pgxpool.Pool) StatsRepository {
	return &statsRepository{db: db}
}

func (r *statsRepository) GetAssignmentStats(ctx context.Context) ([]*models.UserAssignmentStats, error) {
	query := `
		SELECT 
			u.id as user_id,
			u.username,
			u.team_name,
			u.is_active,
			COUNT(pr.id) as assignment_count
		FROM users u
		LEFT JOIN pull_requests pr ON pr.assigned_reviewers @> jsonb_build_array(u.id)
		WHERE u.is_active = true
		GROUP BY u.id, u.username, u.team_name, u.is_active
		ORDER BY assignment_count DESC, u.username
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignment stats: %w", err)
	}
	defer rows.Close()

	var stats []*models.UserAssignmentStats
	for rows.Next() {
		var stat models.UserAssignmentStats
		err := rows.Scan(
			&stat.UserID,
			&stat.Username,
			&stat.TeamName,
			&stat.IsActive,
			&stat.AssignmentCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment stats: %w", err)
		}
		stats = append(stats, &stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignment stats: %w", err)
	}

	return stats, nil
}

func (r *statsRepository) GetPRAssignmentStats(ctx context.Context) (*models.PRAssignmentStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_prs,
			COUNT(CASE WHEN status = 'OPEN' THEN 1 END) as open_prs,
			COUNT(CASE WHEN status = 'MERGED' THEN 1 END) as merged_prs,
			COALESCE(AVG(jsonb_array_length(assigned_reviewers)), 0) as avg_reviewers_per_pr,
			COUNT(*) FILTER (WHERE jsonb_array_length(assigned_reviewers) = 0) as prs_with_no_reviewers,
			COUNT(*) FILTER (WHERE jsonb_array_length(assigned_reviewers) = 1) as prs_with_one_reviewer,
			COUNT(*) FILTER (WHERE jsonb_array_length(assigned_reviewers) = 2) as prs_with_two_reviewers
		FROM pull_requests
	`

	var stats models.PRAssignmentStats
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalPRs,
		&stats.OpenPRs,
		&stats.MergedPRs,
		&stats.AvgReviewersPerPR,
		&stats.PRsWithNoReviewers,
		&stats.PRsWithOneReviewer,
		&stats.PRsWithTwoReviewers,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get PR assignment stats: %w", err)
	}

	return &stats, nil
}

func (r *statsRepository) GetUserAssignmentCount(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM pull_requests
		WHERE assigned_reviewers @> $1
	`

	userJSON, err := json.Marshal([]string{userID})
	if err != nil {
		return 0, fmt.Errorf("failed to marshal user filter: %w", err)
	}

	var count int
	err = r.db.QueryRow(ctx, query, userJSON).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user assignment count: %w", err)
	}

	return count, nil
}
