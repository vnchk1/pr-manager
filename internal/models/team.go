package models

import (
	"time"
)

type Team struct {
	Name    string  `json:"team_name"`
	Members []*User `json:"members"`

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (t *Team) Validate() error {
	if t.Name == "" {
		return ErrInvalidTeamName
	}

	for _, member := range t.Members {
		if err := member.Validate(); err != nil {
			return err
		}

		member.TeamName = t.Name
	}

	return nil
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
