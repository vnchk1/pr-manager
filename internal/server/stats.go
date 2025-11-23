package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type StatsResponse struct {
	UserStats []*UserStatsItem `json:"user_stats"`
	PRStats   *PRStats         `json:"pr_stats"`
	Summary   *Summary         `json:"summary"`
}

type UserStatsItem struct {
	UserID          string `json:"user_id"`
	Username        string `json:"username"`
	TeamName        string `json:"team_name"`
	IsActive        bool   `json:"is_active"`
	AssignmentCount int    `json:"assignment_count"`
}

type PRStats struct {
	TotalPRs            int     `json:"total_prs"`
	OpenPRs             int     `json:"open_prs"`
	MergedPRs           int     `json:"merged_prs"`
	AvgReviewersPerPR   float64 `json:"avg_reviewers_per_pr"`
	PRsWithNoReviewers  int     `json:"prs_with_no_reviewers"`
	PRsWithOneReviewer  int     `json:"prs_with_one_reviewer"`
	PRsWithTwoReviewers int     `json:"prs_with_two_reviewers"`
}

type Summary struct {
	TotalUsers       int    `json:"total_users"`
	ActiveUsers      int    `json:"active_users"`
	TotalAssignments int    `json:"total_assignments"`
	MostAssignedUser string `json:"most_assigned_user,omitempty"`
	MostAssignments  int    `json:"most_assignments,omitempty"`
}

type UserStatsResponse struct {
	UserID          string `json:"user_id"`
	Username        string `json:"username"`
	TeamName        string `json:"team_name"`
	IsActive        bool   `json:"is_active"`
	AssignmentCount int    `json:"assignment_count"`
	Rank            int    `json:"rank,omitempty"`
}

func (s *Server) getStats(c echo.Context) error {
	stats, err := s.service.Stats.GetAssignmentStats(c.Request().Context())
	if err != nil {
		return err
	}

	response := StatsResponse{
		UserStats: make([]*UserStatsItem, len(stats.UserStats)),
		PRStats: &PRStats{
			TotalPRs:            stats.PRStats.TotalPRs,
			OpenPRs:             stats.PRStats.OpenPRs,
			MergedPRs:           stats.PRStats.MergedPRs,
			AvgReviewersPerPR:   stats.PRStats.AvgReviewersPerPR,
			PRsWithNoReviewers:  stats.PRStats.PRsWithNoReviewers,
			PRsWithOneReviewer:  stats.PRStats.PRsWithOneReviewer,
			PRsWithTwoReviewers: stats.PRStats.PRsWithTwoReviewers,
		},
		Summary: &Summary{
			TotalUsers:       stats.Summary.TotalUsers,
			ActiveUsers:      stats.Summary.ActiveUsers,
			TotalAssignments: stats.Summary.TotalAssignments,
			MostAssignedUser: stats.Summary.MostAssignedUser,
			MostAssignments:  stats.Summary.MostAssignments,
		},
	}

	for i, userStat := range stats.UserStats {
		response.UserStats[i] = &UserStatsItem{
			UserID:          userStat.UserID,
			Username:        userStat.Username,
			TeamName:        userStat.TeamName,
			IsActive:        userStat.IsActive,
			AssignmentCount: userStat.AssignmentCount,
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) getUserStats(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
	}

	userStats, err := s.service.Stats.GetUserStats(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	response := UserStatsResponse{
		UserID:          userStats.UserID,
		Username:        userStats.Username,
		TeamName:        userStats.TeamName,
		IsActive:        userStats.IsActive,
		AssignmentCount: userStats.AssignmentCount,
	}

	return c.JSON(http.StatusOK, response)
}
