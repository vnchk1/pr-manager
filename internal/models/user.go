package models

import (
	"time"
)

type User struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (u *User) Validate() error {
	if u.ID == "" {
		return ErrInvalidUserID
	}
	if u.Username == "" {
		return ErrInvalidUsername
	}
	if u.TeamName == "" {
		return ErrInvalidTeamName
	}
	return nil
}

type UserUpdate struct {
	Username *string `json:"username,omitempty"`
	TeamName *string `json:"team_name,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}
