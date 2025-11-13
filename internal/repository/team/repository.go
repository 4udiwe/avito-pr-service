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

	var team entity.Team

	err := r.GetTxManager(ctx).QueryRow(ctx, query, args...).Scan(
		&team.ID,
		&team.CreatedAt,
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

	logrus.Infof("TeamRepository.Create: team created with ID %s", team.ID)
	return team, nil
}

func (r *Repository) GetByName(ctx context.Context, name string) (entity.Team, error) {
	logrus.Infof("TeamRepository.GetByName: getting team by name %s", name)

	query, args, _ := r.Builder.Select("id", "created_at").
		From("team").
		Where("name = ?", name).
		ToSql()

	var team entity.Team

	err := r.GetTxManager(ctx).QueryRow(ctx, query, args...).Scan(
		&team.ID,
		&team.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Warnf("TeamRepository.GetByName: team not found: %s", name)
			return entity.Team{}, repository.ErrTeamNotFound
		}
		logrus.Errorf("TeamRepository.GetByName: failed to get team: %v", err)
		return entity.Team{}, err
	}

	logrus.Infof("TeamRepository.GetByName: team found with ID %s", team.ID)
	return team, nil
}
