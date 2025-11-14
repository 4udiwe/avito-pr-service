package get_team

import (
	"errors"
	"net/http"

	api "github.com/4udiwe/avito-pr-service/internal/api/http"
	"github.com/4udiwe/avito-pr-service/internal/api/http/decorator"
	"github.com/4udiwe/avito-pr-service/internal/dto"
	service "github.com/4udiwe/avito-pr-service/internal/service/team"
	"github.com/labstack/echo/v4"
)

type handler struct {
	s TeamService
}

func New(teamService TeamService) api.Handler {
	return decorator.NewBindAndValidateDerocator(&handler{s: teamService})
}

type Request struct {
	TeamName string `param:"team_name" validate:"required"`
}

func (h *handler) Handle(ctx echo.Context, in Request) error {
	team, err := h.s.GetTeamWithMembers(ctx.Request().Context(), in.TeamName)

	if err != nil {
		var errResponse dto.ErrorResponse

		if errors.Is(err, service.ErrTeamNotFound) {
			errResponse.Error.Code = dto.NOTFOUND
			errResponse.Error.Message = "resource not found"
			return echo.NewHTTPError(http.StatusBadRequest, errResponse)
		}

		errResponse.Error.Message = err.Error()
		return echo.NewHTTPError(http.StatusInternalServerError, errResponse)
	}

	var response dto.Team
	response.FillFromEntity(team)

	return ctx.JSON(http.StatusOK, response)
}
