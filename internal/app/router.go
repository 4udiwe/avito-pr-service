package app

import (
	"net/http"

	"github.com/4udiwe/avito-pr-service/pkg/validator"
	"github.com/labstack/echo/v4"
)

func (app *App) EchoHandler() *echo.Echo {
	if app.echoHandler != nil {
		return app.echoHandler
	}

	handler := echo.New()
	handler.Validator = validator.NewCustomValidator()

	app.configureRouter(handler)

	app.echoHandler = handler
	return app.echoHandler
}

func (app *App) configureRouter(handler *echo.Echo) {
	teamGroup := handler.Group("team")
	{
		teamGroup.POST("/add", app.PostTeamHandler().Handle)
		teamGroup.GET("/get", app.GetTeamHandler().Handle)
		teamGroup.GET("", app.GetTeamsHandler().Handle)
		teamGroup.POST("/deactivate", app.PostDeactivateTeamHandler().Handle)
	}

	userGroup := handler.Group("users")
	{
		userGroup.POST("/setIsActive", app.PostIsUserActiveHandler().Handle)
		userGroup.GET("/getReview", app.GetUserReviewsHandler().Handle)
	}

	pullRequestGroup := handler.Group("pullRequest")
	{
		pullRequestGroup.POST("/create", app.PostPRHandler().Handle)
		pullRequestGroup.POST("/merge", app.PostMergePRHandler().Handle)
		pullRequestGroup.POST("/reassign", app.PostReassignReviewerHandler().Handle)
		pullRequestGroup.POST("/assign", app.PostAssignUserToPRHandler().Handle)
		pullRequestGroup.GET("", app.GetPRsHandler().Handle)
	}

	handler.GET("/stats", app.GetStatsHandler().Handle)

	handler.GET("/health", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
}
