package repo_user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	"github.com/4udiwe/avito-pr-service/pkg/postgres"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) *Repository {
	return &Repository{pg}
}

func (r *Repository) Create(ctx context.Context, ID, name string, teamID uuid.UUID, isActive bool) (entity.User, error) {
	logrus.Infof("UserRepository.Create: creating user with name %s", name)

	query, args, _ := r.Builder.Insert("app_user").
		Columns("id", "name", "team_id", "is_active").
		Values(ID, name, teamID, isActive).
		Suffix("RETURNING created_at").
		ToSql()

	rowUser := RowUser{
		ID:       ID,
		Name:     name,
		TeamID:   teamID,
		IsActive: isActive,
	}

	err := r.GetTxManager(ctx).QueryRow(ctx, query, args...).Scan(
		&rowUser.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return entity.User{}, repository.ErrUserAlreadyExists
			}
		}
		logrus.Errorf("UserRepository.Create: failed to create user: %v", err)
		return entity.User{}, err
	}

	logrus.Infof("UserRepository.Create: user created with ID %s", rowUser.ID)
	return rowUser.ToEntity(), nil
}

func (r *Repository) GetByID(ctx context.Context, ID string) (entity.User, error) {
	logrus.Infof("UserRepository.GetByID: getting user by ID %s", ID)

	query, args, _ := r.Builder.
		Select(
			"u.id",
			"u.name",
			"u.is_active",
			"u.team_id",
			"t.name AS team_name",
			"u.created_at",
		).
		From("app_user AS u").
		LeftJoin("team AS t ON u.team_id = t.id").
		Where("u.id = ?", ID).
		ToSql()

	var row RowUser
	err := r.GetTxManager(ctx).QueryRow(ctx, query, args...).Scan(
		&row.ID,
		&row.Name,
		&row.IsActive,
		&row.TeamID,
		&row.TeamName,
		&row.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, repository.ErrUserNotFound
		}
		logrus.Errorf("UserRepository.GetByID: failed to get user by ID %s: %v", ID, err)
		return entity.User{}, err
	}

	logrus.Infof("UserRepository.GetByID: user found with ID %s", row.ID)
	return row.ToEntity(), nil
}

func (r *Repository) GetByTeamID(ctx context.Context, teamID uuid.UUID) ([]entity.User, error) {
	logrus.Infof("UserRepository.GetByTeamID: getting users by team ID %s", teamID)

	query, args, _ := r.Builder.Select("id", "name", "is_active", "team_id", "created_at").
		From("app_user").
		Where("team_id = ?", teamID).
		ToSql()

	rows, err := r.GetTxManager(ctx).Query(ctx, query, args...)
	if err != nil {
		logrus.Errorf("UserRepository.GetByTeamID: failed to query users by team ID %s: %v", teamID, err)
		return nil, err
	}
	defer rows.Close()

	rowsUsers, err := pgx.CollectRows(rows, pgx.RowToStructByName[RowUser])
	if err != nil {
		logrus.Errorf("UserRepository.GetByTeamID: failed to scan user row for team ID %s: %v", teamID, err)
		return nil, err
	}

	users := lo.Map(rowsUsers, func(r RowUser, _ int) entity.User { return r.ToEntity() })

	logrus.Infof("UserRepository.GetByTeamID: found %d users for team ID %s", len(users), teamID)
	return users, nil
}

func (r *Repository) SetTeamID(ctx context.Context, userID string, teamID uuid.UUID) error {
	logrus.Infof("UserRepository.SetTeamID: setting team ID %s for user ID %s", teamID, userID)

	query, args, _ := r.Builder.Update("app_user").
		Set("team_id", teamID).
		Where("id = ?", userID).
		ToSql()

	cmdTag, err := r.GetTxManager(ctx).Exec(ctx, query, args...)
	if err != nil {
		logrus.Errorf("UserRepository.SetTeamID: failed to set team ID %s for user ID %s: %v", teamID, userID, err)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *Repository) SetActiveStatus(ctx context.Context, userID string, isActive bool) error {
	logrus.Infof("UserRepository.SetActiveStatus: setting isActive=%t for user ID %s", isActive, userID)

	query, args, _ := r.Builder.Update("app_user").
		Set("is_active", isActive).
		Where("id = ?", userID).
		ToSql()

	cmdTag, err := r.GetTxManager(ctx).Exec(ctx, query, args...)
	if err != nil {
		logrus.Errorf("UserRepository.SetActiveStatus: failed to set isActive=%t for user ID %s: %v", isActive, userID, err)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *Repository) GetRandomActiveTeammates(ctx context.Context, teamID uuid.UUID, limit int, excludeIDs ...string) ([]entity.User, error) {
	logrus.Infof("UserRepository.GetRandomActiveTeammates: getting up to %d random active teammates for team ID %s", limit, teamID)

	query, args, _ := r.Builder.
		Select("id", "name", "team_id", "created_at").
		From("app_user").
		Where("team_id = ? AND is_active = TRUE", teamID).
		Where(squirrel.NotEq{"id": excludeIDs}).
		OrderBy("RANDOM()").
		Limit(uint64(limit)).
		ToSql()

	rows, err := r.GetTxManager(ctx).Query(ctx, query, args...)
	if err != nil {
		logrus.Errorf("UserRepository.GetRandomActiveTeammates: failed to query random active teammates: %v", err)
		return nil, err
	}
	defer rows.Close()

	rowsUsers, err := pgx.CollectRows(rows, pgx.RowToStructByName[RowUser])
	if err != nil {
		logrus.Errorf("UserRepository.GetRandomActiveTeammates: failed to scan user row for team ID %s: %v", teamID, err)
		return nil, err
	}

	users := lo.Map(rowsUsers, func(r RowUser, _ int) entity.User { return r.ToEntity() })

	logrus.Infof("UserRepository.GetRandomActiveTeammates: found %d random active teammates", len(users))
	return users, nil
}
