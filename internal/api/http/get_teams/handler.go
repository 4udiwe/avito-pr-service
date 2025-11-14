package get_teams

import (
	"math"
	"net/http"

	api "github.com/4udiwe/avito-pr-service/internal/api/http"
	"github.com/4udiwe/avito-pr-service/internal/api/http/decorator"
	"github.com/4udiwe/avito-pr-service/internal/dto"
	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

const PAGE_NUMBER = 1
const PAGE_SIZE = 10

type handler struct {
	s TeamService
}

func New(teamService TeamService) api.Handler {
	return decorator.NewBindAndValidateDerocator(&handler{s: teamService})
}

type GetAllTeamsRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

type GetAllTeamsResponse struct {
	Teams      []dto.Team `json:"teams"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalItems int        `json:"total_items"`
	TotalPages int        `json:"total_pages"`
}

func (h *handler) Handle(ctx echo.Context, in GetAllTeamsRequest) error {
	if in.Page == 0 {
		in.Page = PAGE_NUMBER
	}

	if in.PageSize <= 0 {
		in.PageSize = PAGE_SIZE
	} else if in.PageSize > 100 {
		in.PageSize = 100
	}

	teams, totalCount, err := h.s.GetAllTeams(ctx.Request().Context(), in.Page, in.PageSize)

	if err != nil {
		var errResponse dto.ErrorResponse
		errResponse.Error.Message = err.Error()
		return echo.NewHTTPError(http.StatusInternalServerError, errResponse)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(in.PageSize)))

	return ctx.JSON(http.StatusOK, GetAllTeamsResponse{
		Teams: lo.Map(teams, func(e entity.Team, _ int) dto.Team {
			t := dto.Team{}
			t.FillFromEntity(e)
			return t
		}),
		Page:       in.Page,
		PageSize:   in.PageSize,
		TotalItems: totalCount,
		TotalPages: totalPages,
	})
}
