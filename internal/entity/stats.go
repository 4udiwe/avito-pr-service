package entity

type Stats struct {
	PullRequests PullRequestStats `json:"pull_request_stats"`
	Users        UserStats        `json:"user_stats"`
	Teams        TeamStats        `json:"team_stats"`
}

type PullRequestStats struct {
	TotalPRs  int64 `json:"total_prs"`
	OpenPRs   int64 `json:"open_prs"`
	MergedPRs int64 `json:"merged_prs"`
}

type UserStats struct {
	MostBusyUsers []UserAssignment `json:"most_busy_users"`
	ActiveUsers   int64            `json:"active_users"`
	InactiveUsers int64            `json:"inactive_users"`
}

type UserAssignment struct {
	UserID      string `json:"user_id"`
	Username    string `json:"user_name"`
	Assignments int64  `json:"assignments"`
}

type TeamStats struct {
	TotalTeams     int64               `json:"total_teams"`
	MostActiveTeam MostActiveTeamStats `json:"most_active_team"`
}

type MostActiveTeamStats struct {
	TeamName string `json:"team_name"`
	PRsCount int64  `json:"team_pr_count"`
}
