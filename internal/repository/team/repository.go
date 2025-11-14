package repo_team

import (
	"context"
	"errors"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	"github.com/4udiwe/avito-pr-service/pkg/postgres"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) *Repository {
	return &Repository{pg}
}

func (r *Repository) Create(ctx context.Context, name string) (entity.Team, error) {
	logrus.Infof("TeamRepository.Create: creating team with name %s", name)

	query, args, _ := r.Builder.Insert("team").
		Columns("name").
		Values(name).
		Suffix("RETURNING id, created_at").
		ToSql()

	rowTeam := RowTeam{
		Name: name,
	}

	err := r.GetTxManager(ctx).QueryRow(ctx, query, args...).Scan(
		&rowTeam.ID,
		&rowTeam.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				logrus.Warnf("TeamRepository.Create: team already exists: %s", name)
				return entity.Team{}, repository.ErrTeamAlreadyExists
			}
		}
		logrus.Errorf("TeamRepository.Create: failed to create team: %v", err)
		return entity.Team{}, err
	}

	logrus.Infof("TeamRepository.Create: team created: %s", name)
	return rowTeam.ToEntity(), nil
}

func (r *Repository) GetByName(ctx context.Context, name string) (entity.Team, error) {
	logrus.Infof("TeamRepository.GetByName: getting team by name %s", name)

	query, args, _ := r.Builder.Select("id", "created_at").
		From("team").
		Where("name = ?", name).
		ToSql()

	rowTeam := RowTeam{
		Name: name,
	}

	err := r.GetTxManager(ctx).QueryRow(ctx, query, args...).Scan(
		&rowTeam.ID,
		&rowTeam.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Warnf("TeamRepository.GetByName: team not found: %s", name)
			return entity.Team{}, repository.ErrTeamNotFound
		}
		logrus.Errorf("TeamRepository.GetByName: failed to get team: %v", err)
		return entity.Team{}, err
	}

	logrus.Infof("TeamRepository.GetByName: team found with ID %s", rowTeam.ID)
	return rowTeam.ToEntity(), nil
}

func (r *Repository) GetAll(
	ctx context.Context,
	limit int,
	offset int,
) (teams []entity.Team, total int, err error) {
	logrus.Info("TeamRepository.GetAll called")

	query, args, _ := r.Builder.
		Select(
			"id",
			"name",
			"created_at",
		).
		From("team").
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	rows, err := r.GetTxManager(ctx).Query(ctx, query, args...)
	if err != nil {
		logrus.Error("TeamRepository.GetAll error: ", err)
		return nil, 0, repository.ErrCannotFetchTeams
	}
	defer rows.Close()

	for rows.Next() {
		var t entity.Team

		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.CreatedAt,
		); err != nil {
			logrus.Error("TeamRepository.GetAll scan error: ", err)
			return nil, 0, repository.ErrCannotFetchTeams
		}

		teams = append(teams, t)
	}

	// Get total count
	countQuery, countArgs, _ := r.Builder.
		Select("COUNT(*)").
		From("team").
		ToSql()

	if err := r.GetTxManager(ctx).QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		logrus.Error("TeamRepository.GetAll - failed to get total count: ", err)
		return nil, 0, repository.ErrCannotFetchTeams
	}

	logrus.Infof("TeamRepository.GetAll success: count=%d", len(teams))
	return teams, total, nil
}
