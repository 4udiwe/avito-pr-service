package user

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrCannotSetUserStatus  = errors.New("cannot set user status")
	ErrCannotGetUserReviews = errors.New("cannot get user reviews")
)
