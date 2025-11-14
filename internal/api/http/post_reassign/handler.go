package post_reassign

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

type Request dto.PostPullRequestReassignJSONRequestBody

type Response struct {
	PR            dto.PullRequest `json:"pr"`
	NewReviewerID string          `json:"replaced_by"`
}

func (h *handler) Handle(ctx echo.Context, in Request) error {
	PR, newReviewerID, err := h.s.ReassignReviewer(ctx.Request().Context(), in.PullRequestId, in.OldUserId)

	if err != nil {
		var errResponse dto.ErrorResponse

		if errors.Is(err, service.ErrPRNotFound) || errors.Is(err, service.ErrReviewerNotFound) {
			errResponse.Error.Code = dto.NOTFOUND
			errResponse.Error.Message = "resource not found"
			return echo.NewHTTPError(http.StatusNotFound, errResponse)
		}
		if errors.Is(err, service.ErrCannotReassignReviewerForMergedPR) {
			errResponse.Error.Code = dto.PRMERGED
			errResponse.Error.Message = "cannot reassign on merged PR"
			return echo.NewHTTPError(http.StatusNotFound, errResponse)
		}

		errResponse.Error.Message = err.Error()
		return echo.NewHTTPError(http.StatusInternalServerError, errResponse)
	}

	response := Response{
		NewReviewerID: newReviewerID,
	}
	response.PR.FillFromEntity(PR)

	return ctx.JSON(http.StatusOK, response)
}
