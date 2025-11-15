package post_deactivate_team

import (
	"net/http"

	api "github.com/4udiwe/avito-pr-service/internal/api/http"
	"github.com/4udiwe/avito-pr-service/internal/api/http/decorator"
	"github.com/4udiwe/avito-pr-service/internal/dto"
	"github.com/labstack/echo/v4"
)

type handler struct {
	s TeamService
}

func New(teamService TeamService) api.Handler {
	return decorator.NewBindAndValidateDerocator(&handler{s: teamService})
}

type Request struct {
	TeamName string `json:"team_name" validate:"required"`
}

func (h *handler) Handle(ctx echo.Context, in Request) error {
	err := h.s.DeactivateTeamAndReassignPRs(ctx.Request().Context(), in.TeamName)

	if err != nil {
		var errResponse dto.ErrorResponse
		errResponse.Error.Message = err.Error()
		return echo.NewHTTPError(http.StatusInternalServerError, errResponse)
	}

	return ctx.NoContent(http.StatusOK)
}
