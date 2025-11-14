package pr

import "errors"

var (
	ErrPRAlreadyExists                   = errors.New("PR already exists")
	ErrAuthorNotFound                    = errors.New("author not found")
	ErrReviewerNotFound                  = errors.New("reviewer not found")
	ErrReviewerAlreadyAssigned           = errors.New("reviewer already assigned to PR")
	ErrPRNotFound                        = errors.New("PR not found")
	ErrCannotCreatePR                    = errors.New("cannot create PR")
	ErrCannotReassignReviewerForMergedPR = errors.New("cannot reassign reviewer for merged PR")
	ErrCannotMergePR                     = errors.New("cannot merge PR")
	ErrStatusNotFound                    = errors.New("status not found")
	ErrNoMoreReviewersToReassign         = errors.New("no more reviewers to reassign")
)
