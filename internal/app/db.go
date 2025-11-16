package app

import (
	repo_pr "github.com/4udiwe/avito-pr-service/internal/repository/pr"
	repo_stats "github.com/4udiwe/avito-pr-service/internal/repository/stats"
	repo_team "github.com/4udiwe/avito-pr-service/internal/repository/team"
	repo_user "github.com/4udiwe/avito-pr-service/internal/repository/user"
	"github.com/4udiwe/avito-pr-service/pkg/postgres"
)

func (app *App) Postgres() *postgres.Postgres {
	return app.postgres
}

func (app *App) UserRepo() *repo_user.Repository {
	if app.userRepo != nil {
		return app.userRepo
	}
	app.userRepo = repo_user.New(app.Postgres())
	return app.userRepo
}

func (app *App) TeamRepo() *repo_team.Repository {
	if app.teamRepo != nil {
		return app.teamRepo
	}
	app.teamRepo = repo_team.New(app.Postgres())
	return app.teamRepo
}

func (app *App) PRRepo() *repo_pr.Repository {
	if app.prRepo != nil {
		return app.prRepo
	}
	app.prRepo = repo_pr.New(app.Postgres())
	return app.prRepo
}

func (app *App) StatsRepo() *repo_stats.Repository {
	if app.statsRepo != nil {
		return app.statsRepo
	}
	app.statsRepo = repo_stats.New(app.Postgres())
	return app.statsRepo
}
