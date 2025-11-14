package pr

import "errors"

var (
	ErrPRAlreadyExists = errors.New("PR already exists")
	ErrPRNotFound      = errors.New("PR not found")
	ErrAuthorNotFound  = errors.New("author not found")
	ErrCannotFetchPRs  = errors.New("cannot fetch PRs")

	ErrCannotCreatePR = errors.New("cannot create PR")
	ErrCannotMergePR  = errors.New("cannot merge PR")

	ErrStatusNotFound = errors.New("status not found")

	ErrReviewerNotFound                  = errors.New("reviewer not found")
	ErrCannotAssignReviewer              = errors.New("cannot assign reviewer")
	ErrCannotReassignReviewerForMergedPR = errors.New("cannot reassign reviewer for merged PR")
	ErrReviewerAlreadyAssigned           = errors.New("reviewer already assigned to PR")
	ErrNoMoreReviewersToReassign         = errors.New("no more reviewers to reassign")
	ErrPRAlreadyHas2Reviewers            = errors.New("PR already has 2 reviewers")
)
