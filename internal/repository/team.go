package repository

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type teamRepository struct {
	db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(ctx context.Context, team *models.Team) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	teamQuery := `
		INSERT INTO teams (name)
		VALUES ($1)
		ON CONFLICT (name) DO NOTHING
	`

	result, err := tx.Exec(ctx, teamQuery, team.Name)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrTeamExists
	}

	userQuery := `
		INSERT INTO users (id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active,
			updated_at = CURRENT_TIMESTAMP
	`

	for _, member := range team.Members {
		_, err := tx.Exec(ctx, userQuery, member.ID, member.Username, team.Name, member.IsActive)
		if err != nil {
			return fmt.Errorf("failed to create user %s: %w", member.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *teamRepository) GetByName(ctx context.Context, teamName string) (*models.Team, error) {

	teamQuery := `
		SELECT name, created_at, updated_at
		FROM teams
		WHERE name = $1
	`

	var team models.Team
	err := r.db.QueryRow(ctx, teamQuery, teamName).Scan(
		&team.Name,
		&team.CreatedAt,
		&team.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	membersQuery := `
		SELECT id, username, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY username
	`

	rows, err := r.db.Query(ctx, membersQuery, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var member models.User
		err := rows.Scan(
			&member.ID,
			&member.Username,
			&member.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		member.TeamName = teamName
		team.Members = append(team.Members, &member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team members: %w", err)
	}

	return &team, nil
}

func (r *teamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}

	return exists, nil
}

func (r *teamRepository) Update(ctx context.Context, team *models.Team) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	userQuery := `
		INSERT INTO users (id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active,
			updated_at = CURRENT_TIMESTAMP
	`

	for _, member := range team.Members {
		_, err := tx.Exec(ctx, userQuery, member.ID, member.Username, team.Name, member.IsActive)
		if err != nil {
			return fmt.Errorf("failed to update user %s: %w", member.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
