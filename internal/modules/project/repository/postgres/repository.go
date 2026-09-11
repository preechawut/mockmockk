package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"mockapi/internal/modules/project/domain"
	pkgpostgres "mockapi/internal/platform/postgres"
)

type Repository struct{ db pkgpostgres.DBTX }

func NewRepository(db pkgpostgres.DBTX) *Repository { return &Repository{db: db} }
func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var p *pgconn.PgError
	if errors.As(err, &p) && (p.Code == "23505" || p.Code == "23503") {
		return domain.ErrConflict
	}
	if err == nil {
		return nil
	}
	return err
}
func scan(row pgx.Row) (domain.Project, error) {
	var p domain.Project
	err := row.Scan(&p.ID, &p.Slug, &p.Name, &p.CreatedAt)
	return p, mapError(err)
}
func (r *Repository) Create(ctx context.Context, p domain.Project) (domain.Project, error) {
	return scan(r.db.QueryRow(ctx, `INSERT INTO projects(id,slug,name) VALUES($1,$2,$3) RETURNING id,slug,name,created_at`, p.ID, p.Slug, p.Name))
}
func (r *Repository) Get(ctx context.Context, id string) (domain.Project, error) {
	return scan(r.db.QueryRow(ctx, `SELECT id,slug,name,created_at FROM projects WHERE id=$1`, id))
}
func (r *Repository) Rename(ctx context.Context, id, name string) (domain.Project, error) {
	return scan(r.db.QueryRow(ctx, `UPDATE projects SET name=$2 WHERE id=$1 RETURNING id,slug,name,created_at`, id, name))
}
func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM projects WHERE id=$1`, id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *Repository) List(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.db.Query(ctx, `SELECT id,slug,name,created_at FROM projects ORDER BY created_at,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Project{}
	for rows.Next() {
		p, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (domain.Project, error) {
	return scan(r.db.QueryRow(ctx, `SELECT id,slug,name,created_at FROM projects WHERE slug=$1`, slug))
}
