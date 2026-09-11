package app

import (
	"context"
	projectpostgres "mockapi/internal/modules/project/repository/postgres"
)

type projectLookup struct{ repo *projectpostgres.Repository }

func (p projectLookup) Resolve(ctx context.Context, ref string, slug bool) (string, error) {
	if slug {
		project, err := p.repo.GetBySlug(ctx, ref)
		return project.ID, err
	}
	project, err := p.repo.Get(ctx, ref)
	return project.ID, err
}
