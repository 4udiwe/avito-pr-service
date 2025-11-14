package repository

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")

	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrTeamNotFound      = errors.New("team not found")
	ErrCannotFetchTeams  = errors.New("cannot fetch teams")

	ErrPRNotFound              = errors.New("pull request not found")
	ErrPRAlreadyExists         = errors.New("pull request already exists")
	ErrReviewerAlreadyAssigned = errors.New("reviewer already assigned to this pull request")
	ErrReviewerNotFound        = errors.New("reviewer not found")
	ErrAuthorNotFound          = errors.New("author not found")
	ErrCannotFetchPRs          = errors.New("cannot fetch PRs")

	ErrStatusNotFound = errors.New("status not found")
)
