package post_assign

import (
	"errors"
	"net/http"

	api "github.com/4udiwe/avito-pr-service/internal/api/http"
	"github.com/4udiwe/avito-pr-service/internal/api/http/decorator"
	"github.com/4udiwe/avito-pr-service/internal/dto"
	service "github.com/4udiwe/avito-pr-service/internal/service/pr"
	"github.com/labstack/echo/v4"
)

type handler struct {
	s PRService
}

func New(PRService PRService) api.Handler {
	return decorator.NewBindAndValidateDerocator(&handler{s: PRService})
}

type Request struct {
	PullRequestID string `json:"pull_request_id"`
	NewReviewerID string `json:"new_reviewer_id"`
}

func (h *handler) Handle(ctx echo.Context, in Request) error {
	PR, err := h.s.AssignReviewer(ctx.Request().Context(), in.PullRequestID, in.NewReviewerID)

	if err != nil {
		var errResponse dto.ErrorResponse

		if errors.Is(err, service.ErrPRNotFound) || errors.Is(err, service.ErrReviewerNotFound) {
			errResponse.Error.Code = dto.NOTFOUND
			errResponse.Error.Message = "resource not found"
			return echo.NewHTTPError(http.StatusNotFound, errResponse)
		}
		if errors.Is(err, service.ErrPRAlreadyHas2Reviewers) {
			errResponse.Error.Code = dto.NOTASSIGNED
			errResponse.Error.Message = err.Error()
			return echo.NewHTTPError(http.StatusNotFound, errResponse)
		}

		errResponse.Error.Message = err.Error()
		return echo.NewHTTPError(http.StatusInternalServerError, errResponse)
	}

	response := dto.PullRequest{}
	response.FillFromEntity(PR)

	return ctx.JSON(http.StatusOK, response)
}
