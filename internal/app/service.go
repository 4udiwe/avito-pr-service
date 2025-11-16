package app

import (
	"github.com/4udiwe/avito-pr-service/internal/service/pr"
	"github.com/4udiwe/avito-pr-service/internal/service/stats"
	"github.com/4udiwe/avito-pr-service/internal/service/team"
	"github.com/4udiwe/avito-pr-service/internal/service/user"
)

func (app *App) TeamService() *team.Service {
	if app.teamService != nil {
		return app.teamService
	}
	app.teamService = team.New(app.UserRepo(), app.TeamRepo(), app.PRRepo(), app.Postgres())
	return app.teamService
}

func (app *App) PRService() *pr.Service {
	if app.prService != nil {
		return app.prService
	}
	app.prService = pr.New(app.PRRepo(), app.UserRepo(), app.Postgres())
	return app.prService
}

func (app *App) UserService() *user.Service {
	if app.userService != nil {
		return app.userService
	}
	app.userService = user.New(app.UserRepo(), app.PRRepo(), app.Postgres())
	return app.userService
}

func (app *App) StatsService() *stats.Service {
	if app.statsService != nil {
		return app.statsService
	}
	app.statsService = stats.New(app.StatsRepo())
	return app.statsService
}
