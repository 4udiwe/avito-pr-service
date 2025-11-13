package repository

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")

	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrTeamNotFound      = errors.New("team not found")

	ErrPRNotFound = errors.New("pull request not found")
)
