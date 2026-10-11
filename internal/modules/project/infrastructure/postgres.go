package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"mockapi/internal/modules/project/domain"
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
		return apperror.New(apperror.NotFound, "project not found")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "23503") {
		return apperror.New(apperror.Conflict, "project slug already exists")
	}
	return err
}

func (r *Repository) Create(ctx context.Context, p domain.Project) (domain.Project, error) {
	const query = `
		INSERT INTO projects (id, slug, name, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, slug, name, created_at
	`
	err := r.db.QueryRowContext(ctx, query, p.ID, p.Slug, p.Name, p.CreatedAt).Scan(
		&p.ID,
		&p.Slug,
		&p.Name,
		&p.CreatedAt,
	)
	if err != nil {
		return domain.Project{}, mapError(err)
	}
	return p, nil
}

func (r *Repository) Get(ctx context.Context, id string) (domain.Project, error) {
	const query = `
		SELECT id, slug, name, created_at
		FROM projects
		WHERE id = $1
	`
	var p domain.Project
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.Slug,
		&p.Name,
		&p.CreatedAt,
	)
	if err != nil {
		return domain.Project{}, mapError(err)
	}
	return p, nil
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (domain.Project, error) {
	const query = `
		SELECT id, slug, name, created_at
		FROM projects
		WHERE slug = $1
	`
	var p domain.Project
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&p.ID,
		&p.Slug,
		&p.Name,
		&p.CreatedAt,
	)
	if err != nil {
		return domain.Project{}, mapError(err)
	}
	return p, nil
}

func (r *Repository) Rename(ctx context.Context, id, name string) (domain.Project, error) {
	const query = `
		UPDATE projects
		SET name = $2
		WHERE id = $1
		RETURNING id, slug, name, created_at
	`
	var p domain.Project
	err := r.db.QueryRowContext(ctx, query, id, name).Scan(
		&p.ID,
		&p.Slug,
		&p.Name,
		&p.CreatedAt,
	)
	if err != nil {
		return domain.Project{}, mapError(err)
	}
	return p, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM projects WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return mapError(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperror.New(apperror.NotFound, "project not found")
	}
	return nil
}

func (r *Repository) List(ctx context.Context) ([]domain.Project, error) {
	const query = `
		SELECT id, slug, name, created_at
		FROM projects
		ORDER BY created_at, id
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]domain.Project, 0)
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.Slug, &p.Name, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}
