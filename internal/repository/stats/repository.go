package repo_stats

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/4udiwe/avito-pr-service/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
)

type Repository struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) *Repository {
	return &Repository{pg}
}

func (r *Repository) GetStats(ctx context.Context) (*entity.Stats, error) {
	stats := &entity.Stats{}

	queryPRs, args, _ := r.Builder.
		Select(
			"COUNT(*) AS total_prs",
			// считаем OPEN через JOIN с pr_status
			"COUNT(*) FILTER (WHERE ps.name = 'OPEN') AS open_prs",
			// считаем MERGED точно так же
			"COUNT(*) FILTER (WHERE ps.name = 'MERGED') AS merged_prs",
		).
		From("pr AS p").
		Join("pr_status AS ps ON p.status_id = ps.id").
		ToSql()

	rows, err := r.GetTxManager(ctx).Query(ctx, queryPRs, args...)
	if err != nil {
		return nil, err
	}

	rowRPStats, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RowPullRequestStats])
	if err != nil {
		return nil, err
	}
	stats.PullRequests = *rowRPStats.ToEntity()

	queryUsers, args, _ := r.Builder.
		Select(
			"COUNT(*) FILTER (WHERE is_active=TRUE) AS active_users",
			"COUNT(*) FILTER (WHERE is_active=FALSE) AS inactive_users",
		).
		From("app_user").
		ToSql()

	rows, err = r.GetTxManager(ctx).Query(ctx, queryUsers, args...)
	if err != nil {
		return nil, err
	}

	rowUserStats, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RowUserStats])
	if err != nil {
		return nil, err
	}

	stats.Users = *rowUserStats.ToEntity()

	queryTopUsers, args, _ := r.Builder.
		Select(
			"u.id AS user_id",
			"u.name AS user_name",
			"COUNT(prr.id) AS assignments",
		).
		From("app_user AS u").
		Join("pr_reviewer AS prr ON prr.reviewer_id = u.id").
		GroupBy("u.id, u.name").
		OrderBy("assignments DESC").
		Limit(5).
		ToSql()

	rows, err = r.GetTxManager(ctx).Query(ctx, queryTopUsers, args...)
	if err != nil {
		return nil, err
	}

	rowsUserAssignment, err := pgx.CollectRows(rows, pgx.RowToStructByName[RowUserAssignment])
	if err != nil {
		return nil, err
	}

	stats.Users.MostBusyUsers = lo.Map(rowsUserAssignment, func(r RowUserAssignment, _ int) entity.UserAssignment { return *r.ToEntity() })

	queryTeams, args, _ := r.Builder.
		Select("COUNT(*) AS total_teams").
		From("team").
		ToSql()

	err = r.GetTxManager(ctx).QueryRow(ctx, queryTeams, args...).Scan(&stats.Teams.TotalTeams)
	if err != nil {
		return nil, err
	}

	queryActiveTeam, args, _ := r.Builder.
		Select("t.name AS team_name", "COUNT(pr.id) AS pr_count").
		From("team AS t").
		Join("app_user AS u ON u.team_id = t.id").
		Join("pr AS pr ON pr.author_id = u.id").
		GroupBy("t.id").
		OrderBy("pr_count DESC").
		Limit(1).
		ToSql()

	rows, err = r.GetTxManager(ctx).Query(ctx, queryActiveTeam, args...)
	if err != nil {
		return nil, err
	}

	rowActiveTeam, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[RowMostActiveTeam])
	if err != nil {
		return nil, err
	}

	stats.Teams.MostActiveTeam = *rowActiveTeam.ToEntity()

	return stats, nil
}
