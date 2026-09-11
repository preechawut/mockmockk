package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/modules/endpoint/service"
	pkgpostgres "mockapi/internal/platform/postgres"
)

type DBTX = pkgpostgres.DBTX

type Repository struct {
	db DBTX
}

var _ service.Repository = (*Repository)(nil)

func NewRepository(db DBTX) *Repository {
	return &Repository{db: db}
}

const columns = "id, project_id, method, path, status, response, created_at"

func scanEndpoint(row pgx.Row) (domain.Endpoint, error) {
	var m domain.Endpoint
	var respBytes []byte
	err := row.Scan(&m.ID, &m.ProjectID, &m.Method, &m.Path, &m.Status, &respBytes, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Endpoint{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Endpoint{}, fmt.Errorf("scan mockmock: %w", err)
	}
	m.Response = respBytes
	return m, nil
}

func (r *Repository) Create(ctx context.Context, m domain.Endpoint) (domain.Endpoint, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO mockmocks (id, project_id, method, path, status, response, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING `+columns,
		m.ID, m.ProjectID, m.Method, m.Path, m.Status, []byte(m.Response), m.CreatedAt)

	res, err := scanEndpoint(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Endpoint{}, domain.ErrConflict
		}
		return domain.Endpoint{}, fmt.Errorf("insert mockmock: %w", err)
	}
	return res, nil
}

func (r *Repository) List(ctx context.Context) ([]domain.Endpoint, error) {
	rows, err := r.db.Query(ctx, `SELECT `+columns+` FROM mockmocks ORDER BY path, method`)
	if err != nil {
		return nil, fmt.Errorf("query mockmocks: %w", err)
	}
	defer rows.Close()

	var result []domain.Endpoint
	for rows.Next() {
		m, err := scanEndpoint(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mockmocks: %w", err)
	}
	return result, nil
}

func (r *Repository) ListByProject(ctx context.Context, projectID string) ([]domain.Endpoint, error) {
	rows, err := r.db.Query(ctx, `SELECT `+columns+` FROM mockmocks WHERE project_id = $1 ORDER BY path, method`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query project mockmocks: %w", err)
	}
	defer rows.Close()

	var result []domain.Endpoint
	for rows.Next() {
		m, err := scanEndpoint(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project mockmocks: %w", err)
	}
	return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Endpoint, error) {
	m, err := scanEndpoint(r.db.QueryRow(ctx, `SELECT `+columns+` FROM mockmocks WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Endpoint{}, domain.ErrNotFound
		}
		return domain.Endpoint{}, fmt.Errorf("find mockmock by id: %w", err)
	}
	return m, nil
}

func (r *Repository) FindMatch(ctx context.Context, method, path string) (domain.Endpoint, error) {
	m, err := scanEndpoint(r.db.QueryRow(ctx,
		`SELECT `+columns+` FROM mockmocks WHERE method = $1 AND path = $2 AND project_id IS NULL`, method, path))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Endpoint{}, domain.ErrNoMatch
		}
		return domain.Endpoint{}, fmt.Errorf("find mockmock match: %w", err)
	}
	return m, nil
}

func (r *Repository) FindMatchByProject(ctx context.Context, projectID, method, path string) (domain.Endpoint, error) {
	m, err := scanEndpoint(r.db.QueryRow(ctx,
		`SELECT `+columns+` FROM mockmocks WHERE project_id = $1 AND method = $2 AND path = $3`, projectID, method, path))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Endpoint{}, domain.ErrNoMatch
		}
		return domain.Endpoint{}, fmt.Errorf("find project mockmock match: %w", err)
	}
	return m, nil
}

func (r *Repository) Update(ctx context.Context, m domain.Endpoint) (domain.Endpoint, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE mockmocks SET method = $1, path = $2, status = $3, response = $4 WHERE id = $5
		 RETURNING `+columns,
		m.Method, m.Path, m.Status, []byte(m.Response), m.ID)

	res, err := scanEndpoint(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Endpoint{}, domain.ErrConflict
		}
		return domain.Endpoint{}, fmt.Errorf("update mockmock: %w", err)
	}
	return res, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM mockmocks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete mockmock: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
