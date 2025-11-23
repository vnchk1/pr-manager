package repository

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestContainer создает тестовую БД в контейнере
func SetupTestContainer(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Запускаем контейнер с PostgreSQL
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Получаем хост и порт контейнера
	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Формируем строку подключения
	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	// Ждем пока БД будет готова принимать подключения
	var db *pgxpool.Pool
	require.Eventually(t, func() bool {
		db, err = pgxpool.New(ctx, connStr)
		return err == nil
	}, 10*time.Second, 1*time.Second, "DB should be ready")

	// Выполняем миграции
	err = runMigrations(ctx, db)
	require.NoError(t, err)

	// Функция очистки
	cleanup := func() {
		db.Close()
		postgresContainer.Terminate(ctx)
	}

	return db, cleanup
}

// runMigrations создает таблицы для тестов
func runMigrations(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS pull_requests (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			author_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			assigned_reviewers JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			merged_at TIMESTAMP WITH TIME ZONE,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			team_name VARCHAR(255) NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS teams (
			name VARCHAR(255) PRIMARY KEY,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func TestPullRequestRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *models.PullRequest
		wantErr bool
		prepare func(ctx context.Context, db *pgxpool.Pool)
		assert  func(ctx context.Context, t *testing.T, db *pgxpool.Pool, pr *models.PullRequest)
	}{
		{
			name: "Happy path",
			input: &models.PullRequest{
				ID:                "pr-1",
				Name:              "Test PR",
				AuthorID:          "user-1",
				Status:            models.StatusOpen,
				AssignedReviewers: []string{"user-2", "user-3"},
			},
			wantErr: false,
			assert: func(ctx context.Context, t *testing.T, db *pgxpool.Pool, pr *models.PullRequest) {
				var (
					id, name, authorID, status string
					reviewersJSON              []byte
				)
				err := db.QueryRow(ctx,
					"SELECT id, name, author_id, status, assigned_reviewers FROM pull_requests WHERE id = $1",
					pr.ID,
				).Scan(&id, &name, &authorID, &status, &reviewersJSON)
				require.NoError(t, err)
				require.Equal(t, "pr-1", id)
				require.Equal(t, "Test PR", name)
				require.Equal(t, "user-1", authorID)
				require.Equal(t, "OPEN", status)

				var reviewers []string
				err = json.Unmarshal(reviewersJSON, &reviewers)
				require.NoError(t, err)
				require.Equal(t, []string{"user-2", "user-3"}, reviewers)
			},
		},
		{
			name: "Duplicate ID",
			input: &models.PullRequest{
				ID:                "pr-1",
				Name:              "Test PR",
				AuthorID:          "user-1",
				Status:            models.StatusOpen,
				AssignedReviewers: []string{"user-2"},
			},
			wantErr: true,
			prepare: func(ctx context.Context, db *pgxpool.Pool) {
				reviewersJSON, _ := json.Marshal([]string{"user-2"})
				_, err := db.Exec(ctx,
					"INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers) VALUES ($1, $2, $3, $4, $5)",
					"pr-1", "Existing PR", "user-1", "OPEN", reviewersJSON,
				)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			testDB, cleanup := SetupTestContainer(t)
			defer cleanup()

			if tt.prepare != nil {
				tt.prepare(ctx, testDB)
			}

			repo := NewPullRequestRepository(testDB)
			err := repo.Create(ctx, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.assert != nil {
				tt.assert(ctx, t, testDB, tt.input)
			}
		})
	}
}

func TestPullRequestRepository_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		prID    string
		want    *models.PullRequest
		wantErr bool
		prepare func(ctx context.Context, db *pgxpool.Pool)
	}{
		{
			name: "Happy path",
			prID: "pr-1",
			want: &models.PullRequest{
				ID:                "pr-1",
				Name:              "Test PR",
				AuthorID:          "user-1",
				Status:            models.StatusOpen,
				AssignedReviewers: []string{"user-2", "user-3"},
			},
			wantErr: false,
			prepare: func(ctx context.Context, db *pgxpool.Pool) {
				reviewersJSON, _ := json.Marshal([]string{"user-2", "user-3"})
				_, err := db.Exec(ctx,
					"INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers) VALUES ($1, $2, $3, $4, $5)",
					"pr-1", "Test PR", "user-1", "OPEN", reviewersJSON,
				)
				require.NoError(t, err)
			},
		},
		{
			name:    "Not found",
			prID:    "pr-999",
			want:    nil,
			wantErr: true,
			prepare: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			testDB, cleanup := SetupTestContainer(t)
			defer cleanup()

			if tt.prepare != nil {
				tt.prepare(ctx, testDB)
			}

			repo := NewPullRequestRepository(testDB)
			result, err := repo.GetByID(ctx, tt.prID)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want.ID, result.ID)
				require.Equal(t, tt.want.Name, result.Name)
				require.Equal(t, tt.want.AuthorID, result.AuthorID)
				require.Equal(t, tt.want.Status, result.Status)
				require.Equal(t, tt.want.AssignedReviewers, result.AssignedReviewers)
			}
		})
	}
}

func TestPullRequestRepository_GetByAuthor(t *testing.T) {
	tests := []struct {
		name     string
		authorID string
		want     []*models.PullRequest
		wantErr  bool
		prepare  func(ctx context.Context, db *pgxpool.Pool)
	}{
		{
			name:     "Happy path - multiple PRs",
			authorID: "user-1",
			want: []*models.PullRequest{
				{
					ID:                "pr-2",
					Name:              "Second PR",
					AuthorID:          "user-1",
					Status:            models.StatusMerged,
					AssignedReviewers: []string{"user-3"},
				},
				{
					ID:                "pr-1",
					Name:              "First PR",
					AuthorID:          "user-1",
					Status:            models.StatusOpen,
					AssignedReviewers: []string{"user-2"},
				},
			},
			wantErr: false,
			prepare: func(ctx context.Context, db *pgxpool.Pool) {
				// Вставляем PR в обратном порядке для проверки сортировки
				reviewers1, _ := json.Marshal([]string{"user-2"})
				reviewers2, _ := json.Marshal([]string{"user-3"})

				_, err := db.Exec(ctx,
					"INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
					"pr-1", "First PR", "user-1", "OPEN", reviewers1, time.Now().Add(-2*time.Hour),
				)
				require.NoError(t, err)

				_, err = db.Exec(ctx,
					"INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
					"pr-2", "Second PR", "user-1", "MERGED", reviewers2, time.Now().Add(-1*time.Hour),
				)
				require.NoError(t, err)

				// PR другого автора
				_, err = db.Exec(ctx,
					"INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers) VALUES ($1, $2, $3, $4, $5)",
					"pr-3", "Other PR", "user-2", "OPEN", reviewers1,
				)
				require.NoError(t, err)
			},
		},
		{
			name:     "No PRs for author",
			authorID: "user-999",
			want:     []*models.PullRequest{},
			wantErr:  false,
			prepare:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			testDB, cleanup := SetupTestContainer(t)
			defer cleanup()

			if tt.prepare != nil {
				tt.prepare(ctx, testDB)
			}

			repo := NewPullRequestRepository(testDB)
			result, err := repo.GetByAuthor(ctx, tt.authorID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, result, len(tt.want))

				for i, expected := range tt.want {
					require.Equal(t, expected.ID, result[i].ID)
					require.Equal(t, expected.Name, result[i].Name)
					require.Equal(t, expected.AuthorID, result[i].AuthorID)
					require.Equal(t, expected.Status, result[i].Status)
					require.Equal(t, expected.AssignedReviewers, result[i].AssignedReviewers)
				}
			}
		})
	}
}

func TestPullRequestRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		input   *models.PullRequest
		wantErr bool
		prepare func(ctx context.Context, db *pgxpool.Pool)
		assert  func(ctx context.Context, t *testing.T, db *pgxpool.Pool)
	}{
		{
			name: "Happy path",
			input: &models.PullRequest{
				ID:                "pr-1",
				Name:              "Updated PR",
				AuthorID:          "user-1",
				Status:            models.StatusMerged,
				AssignedReviewers: []string{"user-4", "user-5"},
			},
			wantErr: false,
			prepare: func(ctx context.Context, db *pgxpool.Pool) {
				reviewersJSON, _ := json.Marshal([]string{"user-2", "user-3"})
				_, err := db.Exec(ctx,
					"INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers) VALUES ($1, $2, $3, $4, $5)",
					"pr-1", "Original PR", "user-1", "OPEN", reviewersJSON,
				)
				require.NoError(t, err)
			},
			assert: func(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
				var (
					name, status  string
					reviewersJSON []byte
				)
				err := db.QueryRow(ctx,
					"SELECT name, status, assigned_reviewers FROM pull_requests WHERE id = $1",
					"pr-1",
				).Scan(&name, &status, &reviewersJSON)
				require.NoError(t, err)
				require.Equal(t, "Updated PR", name)
				require.Equal(t, "MERGED", status)

				var reviewers []string
				err = json.Unmarshal(reviewersJSON, &reviewers)
				require.NoError(t, err)
				require.Equal(t, []string{"user-4", "user-5"}, reviewers)
			},
		},
		{
			name: "Not found",
			input: &models.PullRequest{
				ID:                "pr-999",
				Name:              "Non-existent PR",
				AuthorID:          "user-1",
				Status:            models.StatusOpen,
				AssignedReviewers: []string{"user-2"},
			},
			wantErr: true,
			prepare: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			testDB, cleanup := SetupTestContainer(t)
			defer cleanup()

			if tt.prepare != nil {
				tt.prepare(ctx, testDB)
			}

			repo := NewPullRequestRepository(testDB)
			err := repo.Update(ctx, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.assert != nil {
				tt.assert(ctx, t, testDB)
			}
		})
	}
}

func TestPullRequestRepository_Merge(t *testing.T) {
	tests := []struct {
		name     string
		prID     string
		mergedAt time.Time
		wantErr  bool
		prepare  func(ctx context.Context, db *pgxpool.Pool)
		assert   func(ctx context.Context, t *testing.T, db *pgxpool.Pool)
	}{
		{
			name:     "Happy path",
			prID:     "pr-1",
			mergedAt: time.Now(),
			wantErr:  false,
			prepare: func(ctx context.Context, db *pgxpool.Pool) {
				reviewersJSON, _ := json.Marshal([]string{"user-2"})
				_, err := db.Exec(ctx,
					"INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers) VALUES ($1, $2, $3, $4, $5)",
					"pr-1", "Test PR", "user-1", "OPEN", reviewersJSON,
				)
				require.NoError(t, err)
			},
			assert: func(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
				var (
					status   string
					mergedAt *time.Time
				)
				err := db.QueryRow(ctx,
					"SELECT status, merged_at FROM pull_requests WHERE id = $1",
					"pr-1",
				).Scan(&status, &mergedAt)
				require.NoError(t, err)
				require.Equal(t, "MERGED", status)
				require.NotNil(t, mergedAt)
			},
		},
		{
			name:     "Already merged - idempotent",
			prID:     "pr-1",
			mergedAt: time.Now(),
			wantErr:  false,
			prepare: func(ctx context.Context, db *pgxpool.Pool) {
				reviewersJSON, _ := json.Marshal([]string{"user-2"})
				mergedTime := time.Now().Add(-1 * time.Hour)
				_, err := db.Exec(ctx,
					"INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers, merged_at) VALUES ($1, $2, $3, $4, $5, $6)",
					"pr-1", "Test PR", "user-1", "MERGED", reviewersJSON, mergedTime,
				)
				require.NoError(t, err)
			},
			assert: func(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
				var (
					status   string
					mergedAt time.Time
				)
				err := db.QueryRow(ctx,
					"SELECT status, merged_at FROM pull_requests WHERE id = $1",
					"pr-1",
				).Scan(&status, &mergedAt)
				require.NoError(t, err)
				require.Equal(t, "MERGED", status)

				require.True(t, mergedAt.Before(time.Now().Add(-30*time.Minute)))
			},
		},
		{
			name:     "Not found",
			prID:     "pr-999",
			mergedAt: time.Now(),
			wantErr:  true,
			prepare:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			testDB, cleanup := SetupTestContainer(t)
			defer cleanup()

			if tt.prepare != nil {
				tt.prepare(ctx, testDB)
			}

			repo := NewPullRequestRepository(testDB)
			err := repo.Merge(ctx, tt.prID, tt.mergedAt)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.assert != nil {
				tt.assert(ctx, t, testDB)
			}
		})
	}
}

func TestPullRequestRepository_GetOpenPRsWithReviewer(t *testing.T) {
	tests := []struct {
		name       string
		reviewerID string
		want       []*models.PullRequest
		wantErr    bool
		prepare    func(ctx context.Context, db *pgxpool.Pool)
	}{
		{
			name:       "Happy path",
			reviewerID: "user-2",
			want: []*models.PullRequest{
				{
					ID:                "pr-1",
					Name:              "Open PR with reviewer",
					AuthorID:          "user-1",
					Status:            models.StatusOpen,
					AssignedReviewers: []string{"user-2", "user-3"},
				},
			},
			wantErr: false,
			prepare: func(ctx context.Context, db *pgxpool.Pool) {

				reviewers1, _ := json.Marshal([]string{"user-2", "user-3"})
				_, err := db.Exec(ctx,
					"INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers) VALUES ($1, $2, $3, $4, $5)",
					"pr-1", "Open PR with reviewer", "user-1", "OPEN", reviewers1,
				)
				require.NoError(t, err)

				reviewers2, _ := json.Marshal([]string{"user-2"})
				_, err = db.Exec(ctx,
					"INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers) VALUES ($1, $2, $3, $4, $5)",
					"pr-2", "Closed PR with reviewer", "user-1", "CLOSED", reviewers2,
				)
				require.NoError(t, err)

				reviewers3, _ := json.Marshal([]string{"user-4"})
				_, err = db.Exec(ctx,
					"INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers) VALUES ($1, $2, $3, $4, $5)",
					"pr-3", "Open PR without reviewer", "user-1", "OPEN", reviewers3,
				)
				require.NoError(t, err)
			},
		},
		{
			name:       "No open PRs for reviewer",
			reviewerID: "user-999",
			want:       []*models.PullRequest{},
			wantErr:    false,
			prepare:    nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			testDB, cleanup := SetupTestContainer(t)
			defer cleanup()

			if tt.prepare != nil {
				tt.prepare(ctx, testDB)
			}

			repo := NewPullRequestRepository(testDB)
			result, err := repo.GetOpenPRsWithReviewer(ctx, tt.reviewerID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, result, len(tt.want))

				for i, expected := range tt.want {
					require.Equal(t, expected.ID, result[i].ID)
					require.Equal(t, expected.Name, result[i].Name)
					require.Equal(t, expected.AuthorID, result[i].AuthorID)
					require.Equal(t, expected.Status, result[i].Status)
					require.Equal(t, expected.AssignedReviewers, result[i].AssignedReviewers)
				}
			}
		})
	}
}
