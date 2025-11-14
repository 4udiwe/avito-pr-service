package get_user_reviews

import (
	"errors"
	"net/http"

	api "github.com/4udiwe/avito-pr-service/internal/api/http"
	"github.com/4udiwe/avito-pr-service/internal/api/http/decorator"
	"github.com/4udiwe/avito-pr-service/internal/dto"
	"github.com/4udiwe/avito-pr-service/internal/entity"
	service "github.com/4udiwe/avito-pr-service/internal/service/user"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

type handler struct {
	s UserService
}

func New(userService UserService) api.Handler {
	return decorator.NewBindAndValidateDerocator(&handler{s: userService})
}

type Request struct {
	UserID string `param:"user_id" validate:"required"`
}

type Response struct {
	UserID string                 `json:"user_id"`
	PRs    []dto.PullRequestShort `json:"pull_requests"`
}

func (h *handler) Handle(ctx echo.Context, in Request) error {
	PRs, err := h.s.GetUserReviews(ctx.Request().Context(), in.UserID)

	if err != nil {
		var errResponse dto.ErrorResponse

		if errors.Is(err, service.ErrUserNotFound) {
			errResponse.Error.Code = dto.NOTFOUND
			errResponse.Error.Message = "resource not found"
			return echo.NewHTTPError(http.StatusNotFound, errResponse)
		}

		errResponse.Error.Message = err.Error()
		return echo.NewHTTPError(http.StatusInternalServerError, errResponse)
	}

	response := Response{
		UserID: in.UserID,
		PRs: lo.Map(PRs, func(e entity.PullRequest, _ int) dto.PullRequestShort {
			return dto.PullRequestShort{
				AuthorId:        e.AuthorID,
				Status:          dto.PullRequestShortStatus(e.Status.Name),
				PullRequestId:   e.ID,
				PullRequestName: e.Title,
			}
		}),
	}

	return ctx.JSON(http.StatusOK, response)
}
