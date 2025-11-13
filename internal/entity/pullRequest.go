package entity

type PRStatus struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type PullRequest struct {
	ID                string `db:"id"`
	Title             string `db:"title"`
	AuthorID          string `db:"author_id"`
	StatusID          int    `db:"status_id"`
	NeedMoreReviewers bool   `db:"need_more_reviewers"`
	CreatedAt         string `db:"created_at"`
	MergedAt          string `db:"merged_at"`
}

type PRReviewer struct {
	ID         string `db:"id"`
	PRID       string `db:"pr_id"`
	ReviewerID string `db:"reviewer_id"`
	AssignedAt string `db:"assigned_at"`
}

type PRWithReviewerIDs struct {
	PullRequest
	ReviewersIDs []string
}


