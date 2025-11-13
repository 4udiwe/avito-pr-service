package team

import "errors"

var (
	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrTeamNotFound      = errors.New("team not found")
	ErrCannotCreateTeam  = errors.New("cannot create team")
	ErrCannotFetchTeam   = errors.New("cannot fetch team")

	ErrUserAlreadyExists = errors.New("user already exists")
)
