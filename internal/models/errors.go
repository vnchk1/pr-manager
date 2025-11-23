package models

import "errors"

var (
	ErrNotFound = errors.New("resource not found")

	ErrInvalidUserID   = errors.New("invalid user id")
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidTeamName = errors.New("invalid team name")
	ErrUserNotActive   = errors.New("user is not active")

	ErrTeamExists = errors.New("team already exists")

	ErrInvalidPRID      = errors.New("invalid pull request id")
	ErrInvalidPRName    = errors.New("invalid pull request name")
	ErrInvalidAuthorID  = errors.New("invalid author id")
	ErrPRExists         = errors.New("pull request already exists")
	ErrPRMerged         = errors.New("pull request is merged")
	ErrNotAssigned      = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate      = errors.New("no active replacement candidate in team")
	ErrTooManyReviewers = errors.New("too many reviewers assigned")
)
