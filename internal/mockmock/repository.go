package mockmock

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

const cols = `id, method, path, status, response, created_at`

func scanRow(scanFn func(dest ...any) error) (*Mockmock, error) {
	m := &Mockmock{}
	if err := scanFn(&m.ID, &m.Method, &m.Path, &m.Status, &m.Response, &m.CreatedAt); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *Postgres) Create(ctx context.Context, m *Mockmock) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO mockmocks (id, method, path, status, response, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		m.ID, m.Method, m.Path, m.Status, []byte(m.Response), m.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert mockmock: %w", err)
	}
	return nil
}

func (r *Postgres) FindAll(ctx context.Context) ([]*Mockmock, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+cols+` FROM mockmocks ORDER BY path, method`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*Mockmock
	for rows.Next() {
		m, err := scanRow(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Postgres) FindByID(ctx context.Context, id string) (*Mockmock, error) {
	m, err := scanRow(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM mockmocks WHERE id = $1`, id).Scan)
	if err != nil {
		return nil, fmt.Errorf("mockmock not found: %w", err)
	}
	return m, nil
}

func (r *Postgres) FindMatch(ctx context.Context, method, path string) (*Mockmock, error) {
	m, err := scanRow(r.pool.QueryRow(ctx,
		`SELECT `+cols+` FROM mockmocks WHERE method = $1 AND path = $2`, method, path).Scan)
	if err != nil {
		return nil, fmt.Errorf("no mockmock for %s %s: %w", method, path, err)
	}
	return m, nil
}

func (r *Postgres) Update(ctx context.Context, m *Mockmock) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE mockmocks SET method = $1, path = $2, status = $3, response = $4 WHERE id = $5`,
		m.Method, m.Path, m.Status, []byte(m.Response), m.ID)
	return err
}

func (r *Postgres) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM mockmocks WHERE id = $1`, id)
	return err
}

var _ Repository = (*Postgres)(nil)
