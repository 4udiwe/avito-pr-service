package get_stats

import (
	"net/http"

	api "github.com/4udiwe/avito-pr-service/internal/api/http"
	"github.com/4udiwe/avito-pr-service/internal/api/http/decorator"
	"github.com/4udiwe/avito-pr-service/internal/dto"
	"github.com/labstack/echo/v4"
)

type handler struct {
	s StatsService
}

func New(StatsService StatsService) api.Handler {
	return decorator.NewBindAndValidateDerocator(&handler{s: StatsService})
}

type Request struct{}

func (h *handler) Handle(ctx echo.Context, in Request) error {

	stats, err := h.s.GetStats(ctx.Request().Context())

	if err != nil {
		var errResponse dto.ErrorResponse
		errResponse.Error.Message = err.Error()
		return echo.NewHTTPError(http.StatusInternalServerError, errResponse)
	}

	return ctx.JSON(http.StatusOK, stats)
}
