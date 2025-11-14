package post_team

import (
	"errors"
	"net/http"

	api "github.com/4udiwe/avito-pr-service/internal/api/http"
	"github.com/4udiwe/avito-pr-service/internal/api/http/decorator"
	"github.com/4udiwe/avito-pr-service/internal/dto"
	"github.com/4udiwe/avito-pr-service/internal/entity"
	service "github.com/4udiwe/avito-pr-service/internal/service/team"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

type handler struct {
	s TeamService
}

func New(teamService TeamService) api.Handler {
	return decorator.NewBindAndValidateDerocator(&handler{s: teamService})
}

type Request dto.PostTeamAddJSONRequestBody

func (h *handler) Handle(ctx echo.Context, in Request) error {
	users := lo.Map(in.Members, func(m dto.TeamMember, _ int) entity.User {
		return entity.User{
			ID:       m.UserId,
			Name:     m.Username,
			IsActive: m.IsActive,
		}
	})

	team, err := h.s.CreateTeamWithUsers(ctx.Request().Context(), in.TeamName, users)

	if err != nil {
		var errResponse dto.ErrorResponse

		if errors.Is(err, service.ErrTeamAlreadyExists) {
			errResponse.Error.Code = dto.TEAMEXISTS
			errResponse.Error.Message = "team_name already exists"
			return echo.NewHTTPError(http.StatusBadRequest, errResponse)
		}

		errResponse.Error.Message = err.Error()
		return echo.NewHTTPError(http.StatusInternalServerError, errResponse)
	}

	var response dto.Team
	response.FillFromEntity(team)

	return ctx.JSON(http.StatusOK, response)
}
