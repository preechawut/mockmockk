package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/shared/apperror"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func mapError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return apperror.New(apperror.NotFound, "endpoint not found")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.New(apperror.Conflict, "endpoint already exists")
	}
	return err
}

func (r *Repository) Create(ctx context.Context, m domain.Endpoint) (domain.Endpoint, error) {
	const query = `
		INSERT INTO mockmocks (id, project_id, method, path, status, response, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, project_id, method, path, status, response, created_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		m.ID,
		m.ProjectID,
		m.Method,
		m.Path,
		m.Status,
		m.Response,
		m.CreatedAt,
	).Scan(
		&m.ID,
		&m.ProjectID,
		&m.Method,
		&m.Path,
		&m.Status,
		&m.Response,
		&m.CreatedAt,
	)
	if err != nil {
		return domain.Endpoint{}, mapError(err)
	}
	return m, nil
}

func (r *Repository) List(ctx context.Context) ([]domain.Endpoint, error) {
	const query = `
		SELECT id, project_id, method, path, status, response, created_at
		FROM mockmocks
		ORDER BY path, method
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query endpoints: %w", err)
	}
	defer rows.Close()

	result := make([]domain.Endpoint, 0)
	for rows.Next() {
		var m domain.Endpoint
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Method, &m.Path, &m.Status, &m.Response, &m.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (r *Repository) ListByProject(ctx context.Context, projectID string) ([]domain.Endpoint, error) {
	const query = `
		SELECT id, project_id, method, path, status, response, created_at
		FROM mockmocks
		WHERE project_id = $1
		ORDER BY path, method
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("query project endpoints: %w", err)
	}
	defer rows.Close()

	result := make([]domain.Endpoint, 0)
	for rows.Next() {
		var m domain.Endpoint
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Method, &m.Path, &m.Status, &m.Response, &m.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Endpoint, error) {
	const query = `
		SELECT id, project_id, method, path, status, response, created_at
		FROM mockmocks
		WHERE id = $1
	`
	var m domain.Endpoint
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID,
		&m.ProjectID,
		&m.Method,
		&m.Path,
		&m.Status,
		&m.Response,
		&m.CreatedAt,
	)
	if err != nil {
		return domain.Endpoint{}, mapError(err)
	}
	return m, nil
}

func (r *Repository) FindMatch(ctx context.Context, method, path string) (domain.Endpoint, error) {
	const query = `
		SELECT id, project_id, method, path, status, response, created_at
		FROM mockmocks
		WHERE method = $1 AND path = $2 AND project_id IS NULL
	`
	var m domain.Endpoint
	err := r.db.QueryRowContext(ctx, query, method, path).Scan(
		&m.ID,
		&m.ProjectID,
		&m.Method,
		&m.Path,
		&m.Status,
		&m.Response,
		&m.CreatedAt,
	)
	if err != nil {
		return domain.Endpoint{}, mapError(err)
	}
	return m, nil
}

func (r *Repository) FindMatchByProject(ctx context.Context, projectID, method, path string) (domain.Endpoint, error) {
	const query = `
		SELECT id, project_id, method, path, status, response, created_at
		FROM mockmocks
		WHERE project_id = $1 AND method = $2 AND path = $3
	`
	var m domain.Endpoint
	err := r.db.QueryRowContext(ctx, query, projectID, method, path).Scan(
		&m.ID,
		&m.ProjectID,
		&m.Method,
		&m.Path,
		&m.Status,
		&m.Response,
		&m.CreatedAt,
	)
	if err != nil {
		return domain.Endpoint{}, mapError(err)
	}
	return m, nil
}

func (r *Repository) Update(ctx context.Context, m domain.Endpoint) (domain.Endpoint, error) {
	const query = `
		UPDATE mockmocks
		SET method = $1, path = $2, status = $3, response = $4
		WHERE id = $5
		RETURNING id, project_id, method, path, status, response, created_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		m.Method,
		m.Path,
		m.Status,
		m.Response,
		m.ID,
	).Scan(
		&m.ID,
		&m.ProjectID,
		&m.Method,
		&m.Path,
		&m.Status,
		&m.Response,
		&m.CreatedAt,
	)
	if err != nil {
		return domain.Endpoint{}, mapError(err)
	}
	return m, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM mockmocks WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return mapError(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperror.New(apperror.NotFound, "endpoint not found")
	}
	return nil
}
