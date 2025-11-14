package get_prs

import (
	"math"
	"net/http"
	"time"

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
	s PRService
}

func New(PRService PRService) api.Handler {
	return decorator.NewBindAndValidateDerocator(&handler{s: PRService})
}

type GetAllPRsRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

type PullRequest struct {
	AssignedReviewers []string              `json:"assigned_reviewers"`
	AuthorId          string                `json:"author_id"`
	CreatedAt         *time.Time            `json:"createdAt"`
	MergedAt          *time.Time            `json:"mergedAt"`
	PullRequestId     string                `json:"pull_request_id"`
	PullRequestName   string                `json:"pull_request_name"`
	Status            dto.PullRequestStatus `json:"status"`
	// custom field
	NeedMoreReviewers bool `json:"need_more_reviewers"`
}

type GetAllPRsResponse struct {
	PRs        []PullRequest `json:"pull_requests"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalItems int           `json:"total_items"`
	TotalPages int           `json:"total_pages"`
}

func (h *handler) Handle(ctx echo.Context, in GetAllPRsRequest) error {
	if in.Page == 0 {
		in.Page = PAGE_NUMBER
	}

	if in.PageSize <= 0 {
		in.PageSize = PAGE_SIZE
	} else if in.PageSize > 100 {
		in.PageSize = 100
	}

	PRs, totalCount, err := h.s.GetAllPRs(ctx.Request().Context(), in.Page, in.PageSize)

	if err != nil {
		var errResponse dto.ErrorResponse
		errResponse.Error.Message = err.Error()
		return echo.NewHTTPError(http.StatusInternalServerError, errResponse)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(in.PageSize)))

	return ctx.JSON(http.StatusOK, GetAllPRsResponse{
		PRs: lo.Map(PRs, func(e entity.PullRequest, _ int) PullRequest {
			return PullRequest{
				AssignedReviewers: e.Reviewers,
				AuthorId:          e.AuthorID,
				CreatedAt:         &e.CreatedAt,
				MergedAt:          e.MergedAt,
				PullRequestId:     e.ID,
				PullRequestName:   e.Title,
				Status:            dto.PullRequestStatus(e.Status.Name),
				NeedMoreReviewers: e.NeedMoreReviewers,
			}
		}),
		Page:       in.Page,
		PageSize:   in.PageSize,
		TotalItems: totalCount,
		TotalPages: totalPages,
	})
}
