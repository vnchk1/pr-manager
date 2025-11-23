package models

import (
	"time"
)

type PullRequestStatus string

const (
	StatusOpen   PullRequestStatus = "OPEN"
	StatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID                string            `json:"pull_request_id"`
	Name              string            `json:"pull_request_name"`
	AuthorID          string            `json:"author_id"`
	Status            PullRequestStatus `json:"status"`
	AssignedReviewers []string          `json:"assigned_reviewers"` // user_ids

	CreatedAt time.Time  `json:"created_at,omitempty"`
	MergedAt  *time.Time `json:"merged_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
}

func (pr *PullRequest) Validate() error {
	if pr.ID == "" {
		return ErrInvalidPRID
	}
	if pr.Name == "" {
		return ErrInvalidPRName
	}
	if pr.AuthorID == "" {
		return ErrInvalidAuthorID
	}
	if len(pr.AssignedReviewers) > 2 {
		return ErrTooManyReviewers
	}
	return nil
}

type PullRequestShort struct {
	ID       string            `json:"pull_request_id"`
	Name     string            `json:"pull_request_name"`
	AuthorID string            `json:"author_id"`
	Status   PullRequestStatus `json:"status"`
}

type PRCreateRequest struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type PRMergeRequest struct {
	ID string `json:"pull_request_id"`
}

type PRReassignRequest struct {
	ID          string `json:"pull_request_id"`
	OldReviewer string `json:"old_user_id"`
}
