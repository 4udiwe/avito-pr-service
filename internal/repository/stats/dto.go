package repo_stats

import "github.com/4udiwe/avito-pr-service/internal/entity"

type RowPullRequestStats struct {
	TotalPRs  int64 `db:"total_prs"`
	OpenPRs   int64 `db:"open_prs"`
	MergedPRs int64 `db:"merged_prs"`
}

type RowUserAssignment struct {
	UserID      string `db:"user_id"`
	Username    string `db:"user_name"`
	Assignments int64  `db:"assignments"`
}

type RowMostActiveTeam struct {
	TeamName string `db:"team_name"`
	PRCount  int64  `db:"pr_count"`
}

type RowUserStats struct {
	ActiveUsers   int64 `db:"active_users"`
	InactiveUsers int64 `db:"inactive_users"`
}

func (r *RowPullRequestStats) ToEntity() *entity.PullRequestStats {
	return &entity.PullRequestStats{
		TotalPRs:  r.TotalPRs,
		OpenPRs:   r.OpenPRs,
		MergedPRs: r.MergedPRs,
	}
}

func (r *RowUserAssignment) ToEntity() *entity.UserAssignment {
	return &entity.UserAssignment{
		UserID:      r.UserID,
		Username:    r.Username,
		Assignments: r.Assignments,
	}
}

func (r *RowMostActiveTeam) ToEntity() *entity.MostActiveTeamStats {
	return &entity.MostActiveTeamStats{
		TeamName: r.TeamName,
		PRsCount: r.PRCount,
	}
}

func (r *RowUserStats) ToEntity() *entity.UserStats {
	return &entity.UserStats{
		ActiveUsers:   r.ActiveUsers,
		InactiveUsers: r.InactiveUsers,
	}
}
