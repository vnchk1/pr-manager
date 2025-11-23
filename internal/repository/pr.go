package repository

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"

	"github.com/pkg/errors"
)

type pullRequestRepository struct {
	db *pgxpool.Pool
}

func NewPullRequestRepository(db *pgxpool.Pool) PullRequestRepository {
	return &pullRequestRepository{db: db}
}

func (r *pullRequestRepository) Create(ctx context.Context, pr *models.PullRequest) error {
	reviewersJSON, err := json.Marshal(pr.AssignedReviewers)
	if err != nil {
		return errors.Wrap(err, "failed to marshal reviewers")
	}

	query := `
		INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = r.db.Exec(ctx, query,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		pr.Status,
		reviewersJSON,
	)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"pull_requests_pkey\" (SQLSTATE 23505)" {
			return models.ErrPRExists
		}
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	return nil
}

func (r *pullRequestRepository) GetByID(ctx context.Context, prID string) (*models.PullRequest, error) {
	query := `
		SELECT id, name, author_id, status, assigned_reviewers, created_at, merged_at, updated_at
		FROM pull_requests
		WHERE id = $1
	`

	var pr models.PullRequest
	var reviewersJSON []byte

	err := r.db.QueryRow(ctx, query, prID).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&reviewersJSON,
		&pr.CreatedAt,
		&pr.MergedAt,
		&pr.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get pull request by id: %w", err)
	}

	if err = json.Unmarshal(reviewersJSON, &pr.AssignedReviewers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reviewers: %w", err)
	}

	return &pr, nil
}

func (r *pullRequestRepository) GetByAuthor(ctx context.Context, authorID string) ([]*models.PullRequest, error) {
	query := `
		SELECT id, name, author_id, status, assigned_reviewers, created_at, merged_at, updated_at
		FROM pull_requests
		WHERE author_id = $1
		ORDER BY created_at DESC
	`

	return r.queryPullRequests(ctx, query, authorID)
}

func (r *pullRequestRepository) GetByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequestShort, error) {
	query := `
		SELECT id, name, author_id, status
		FROM pull_requests
		WHERE assigned_reviewers @> $1
		ORDER BY created_at DESC
	`

	reviewerJSON, err := json.Marshal([]string{reviewerID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal reviewer filter: %w", err)
	}

	rows, err := r.db.Query(ctx, query, reviewerJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to query pull request by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []*models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		err = rows.Scan(
			&pr.ID,
			&pr.Name,
			&pr.AuthorID,
			&pr.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pull request: %w", err)
		}
		prs = append(prs, &pr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pull request: %w", err)
	}

	return prs, nil
}

func (r *pullRequestRepository) Update(ctx context.Context, pr *models.PullRequest) error {
	reviewersJSON, err := json.Marshal(pr.AssignedReviewers)
	if err != nil {
		return fmt.Errorf("failed to marshal reviewers: %w", err)
	}

	query := `
		UPDATE pull_requests
		SET name = $2, status = $3, assigned_reviewers = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		pr.ID,
		pr.Name,
		pr.Status,
		reviewersJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update pull request: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrNotFound
	}

	return nil
}

func (r *pullRequestRepository) Merge(ctx context.Context, prID string, mergedAt time.Time) error {
	query := `
		UPDATE pull_requests
		SET status = 'MERGED', merged_at = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status != 'MERGED'
	`

	result, err := r.db.Exec(ctx, query, prID, mergedAt)
	if err != nil {
		return fmt.Errorf("failed to merge pull request: %w", err)
	}

	if result.RowsAffected() == 0 {

		exists, err := r.Exists(ctx, prID)
		if err != nil {
			return err
		}
		if !exists {
			return models.ErrNotFound
		}

	}

	return nil
}

func (r *pullRequestRepository) Exists(ctx context.Context, prID string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id = $1)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check pull request existence: %w", err)
	}

	return exists, nil
}

func (r *pullRequestRepository) GetOpenPRsWithReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequest, error) {
	query := `
		SELECT id, name, author_id, status, assigned_reviewers, created_at, merged_at, updated_at
		FROM pull_requests
		WHERE status = 'OPEN' AND assigned_reviewers @> $1
		ORDER BY created_at DESC
	`

	reviewerJSON, err := json.Marshal([]string{reviewerID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshall reviewer: %w", err)
	}

	return r.queryPullRequests(ctx, query, reviewerJSON)
}

func (r *pullRequestRepository) queryPullRequests(ctx context.Context, query string, args ...interface{}) ([]*models.PullRequest, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to failed to query pull request: %w", err)
	}
	defer rows.Close()

	var prs []*models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		var reviewersJSON []byte

		err := rows.Scan(
			&pr.ID,
			&pr.Name,
			&pr.AuthorID,
			&pr.Status,
			&reviewersJSON,
			&pr.CreatedAt,
			&pr.MergedAt,
			&pr.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pull request: %w", err)
		}

		if err := json.Unmarshal(reviewersJSON, &pr.AssignedReviewers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal reviewers: %w", err)
		}

		prs = append(prs, &pr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pull requests: %w", err)
	}

	return prs, nil
}
