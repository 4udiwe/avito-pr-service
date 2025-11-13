package repo_pr

import (
	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type RowStatus struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type RowPullRequest struct {
	ID                string `db:"id"`
	Title             string `db:"title"`
	AuthorID          string `db:"author_id"`
	StatusID          int    `db:"status_id"`
	NeedMoreReviewers bool   `db:"need_more_reviewers"`
	CreatedAt         string `db:"created_at"`
	MergedAt          string `db:"merged_at"`
}

type RowPullRequestWithReviewerIDs struct {
	RowPullRequest
	ReviewerIDs []string
}

type RowPRReviewer struct {
	ID         string `db:"id"`
	PRID       string `db:"pr_id"`
	ReviewerID string `db:"reviewer_id"`
	AssignedAt string `db:"assigned_at"`
}

func (r *RowStatus) ToEntity() entity.Status {
	return entity.Status{
		ID:   r.ID,
		Name: entity.PRStatusName(r.Name),
	}
}

func (r *RowPullRequest) ToEntity() entity.PullRequest {
	return entity.PullRequest{
		ID:                r.ID,
		Title:             r.Title,
		AuthorID:          r.AuthorID,
		Status:            entity.Status{ID: r.StatusID},
		NeedMoreReviewers: r.NeedMoreReviewers,
		CreatedAt:         r.CreatedAt,
		MergedAt:          r.MergedAt,
	}
}

func (r *RowPullRequestWithReviewerIDs) ToEntity() entity.PullRequest {
	return entity.PullRequest{
		ID:                r.ID,
		Title:             r.Title,
		AuthorID:          r.AuthorID,
		Status:            entity.Status{ID: r.StatusID},
		NeedMoreReviewers: r.NeedMoreReviewers,
		CreatedAt:         r.CreatedAt,
		MergedAt:          r.MergedAt,
		Reviewers:         r.ReviewerIDs,
	}
}

func (r *RowPRReviewer) ToEntity() entity.PRReviewer {
	return entity.PRReviewer{
		ID:         r.ID,
		PRID:       r.PRID,
		ReviewerID: r.ReviewerID,
		AssignedAt: r.AssignedAt,
	}
}
