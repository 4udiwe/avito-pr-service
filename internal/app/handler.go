package app

import (
	api "github.com/4udiwe/avito-pr-service/internal/api/http"
	"github.com/4udiwe/avito-pr-service/internal/api/http/get_prs"
	"github.com/4udiwe/avito-pr-service/internal/api/http/get_team"
	"github.com/4udiwe/avito-pr-service/internal/api/http/get_teams"
	"github.com/4udiwe/avito-pr-service/internal/api/http/get_user_reviews"
	"github.com/4udiwe/avito-pr-service/internal/api/http/post_assign"
	"github.com/4udiwe/avito-pr-service/internal/api/http/post_deactivate_team"
	"github.com/4udiwe/avito-pr-service/internal/api/http/post_merge"
	"github.com/4udiwe/avito-pr-service/internal/api/http/post_pr"
	"github.com/4udiwe/avito-pr-service/internal/api/http/post_reassign"
	"github.com/4udiwe/avito-pr-service/internal/api/http/post_team"
	"github.com/4udiwe/avito-pr-service/internal/api/http/post_user_is_active"
)

func (app *App) GetPRsHandler() api.Handler {
	if app.getPRsHandler != nil {
		return app.getPRsHandler
	}
	app.getPRsHandler = get_prs.New(app.PRService())
	return app.getPRsHandler
}

func (app *App) GetTeamHandler() api.Handler {
	if app.getTeamHandler != nil {
		return app.getTeamHandler
	}
	app.getTeamHandler = get_team.New(app.TeamService())
	return app.getTeamHandler
}

func (app *App) GetTeamsHandler() api.Handler {
	if app.getTeamsHandler != nil {
		return app.getTeamsHandler
	}
	app.getTeamsHandler = get_teams.New(app.TeamService())
	return app.getTeamsHandler
}

func (app *App) GetUserReviewsHandler() api.Handler {
	if app.getUserReviewsHandler != nil {
		return app.getTeamsHandler
	}
	app.getTeamsHandler = get_user_reviews.New(app.UserService())
	return app.getTeamsHandler
}

func (app *App) PostAssignUserToPRHandler() api.Handler {
	if app.postAssignUserToPRHandler != nil {
		return app.postAssignUserToPRHandler
	}
	app.postAssignUserToPRHandler = post_assign.New(app.PRService())
	return app.postAssignUserToPRHandler
}

func (app *App) PostMergePRHandler() api.Handler {
	if app.postMergePRHandler != nil {
		return app.postMergePRHandler
	}
	app.postMergePRHandler = post_merge.New(app.PRService())
	return app.postMergePRHandler
}

func (app *App) PostPRHandler() api.Handler {
	if app.postPRHandler != nil {
		return app.postPRHandler
	}
	app.postPRHandler = post_pr.New(app.PRService())
	return app.postPRHandler
}

func (app *App) PostReassignReviewerHandler() api.Handler {
	if app.postReassignReviewerHandler != nil {
		return app.postReassignReviewerHandler
	}
	app.postReassignReviewerHandler = post_reassign.New(app.PRService())
	return app.postReassignReviewerHandler
}

func (app *App) PostTeamHandler() api.Handler {
	if app.postTeamHandler != nil {
		return app.postTeamHandler
	}
	app.postTeamHandler = post_team.New(app.TeamService())
	return app.postTeamHandler
}

func (app *App) PostIsUserActiveHandler() api.Handler {
	if app.postIsUserActiveHandler != nil {
		return app.postIsUserActiveHandler
	}
	app.postIsUserActiveHandler = post_user_is_active.New(app.UserService())
	return app.postIsUserActiveHandler
}

func (app *App) PostDeactivateTeamHandler() api.Handler {
	if app.postDeactivateTeamHandler != nil {
		return app.postDeactivateTeamHandler
	}
	app.postDeactivateTeamHandler = post_deactivate_team.New(app.TeamService())
	return app.postDeactivateTeamHandler
}
