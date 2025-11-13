package entity

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
	CreatedAt         string
	MergedAt          string
	Reviewers         []string
}

type PRReviewer struct {
	ID         string
	PRID       string
	ReviewerID string
	AssignedAt string
}
