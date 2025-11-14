package entity

import "time"

type PRStatusName string

const (
	StatusOPEN   PRStatusName = "OPEN"
	StatusMERGED PRStatusName = "MERGED"
)

type Status struct {
	ID   int
	Name PRStatusName
}

type PullRequest struct {
	ID                string
	Title             string
	AuthorID          string
	Status            Status
	NeedMoreReviewers bool
	CreatedAt         time.Time
	MergedAt          *time.Time
	Reviewers         []string
}

type PRReviewer struct {
	ID         string
	PRID       string
	ReviewerID string
	AssignedAt time.Time
}
