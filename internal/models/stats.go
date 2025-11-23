package models

type UserAssignmentStats struct {
	UserID          string `json:"user_id"`
	Username        string `json:"username"`
	TeamName        string `json:"team_name"`
	IsActive        bool   `json:"is_active"`
	AssignmentCount int    `json:"assignment_count"`
}

type PRAssignmentStats struct {
	TotalPRs            int     `json:"total_prs"`
	OpenPRs             int     `json:"open_prs"`
	MergedPRs           int     `json:"merged_prs"`
	AvgReviewersPerPR   float64 `json:"avg_reviewers_per_pr"`
	PRsWithNoReviewers  int     `json:"prs_with_no_reviewers"`
	PRsWithOneReviewer  int     `json:"prs_with_one_reviewer"`
	PRsWithTwoReviewers int     `json:"prs_with_two_reviewers"`
}

type AssignmentStatsResponse struct {
	UserStats []*UserAssignmentStats `json:"user_stats"`
	PRStats   *PRAssignmentStats     `json:"pr_stats"`
	Summary   *StatsSummary          `json:"summary"`
}

type StatsSummary struct {
	TotalUsers       int    `json:"total_users"`
	ActiveUsers      int    `json:"active_users"`
	TotalAssignments int    `json:"total_assignments"`
	MostAssignedUser string `json:"most_assigned_user,omitempty"`
	MostAssignments  int    `json:"most_assignments,omitempty"`
}
