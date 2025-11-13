package repo_user

import (
	"context"
	"errors"
	"fmt"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	"github.com/4udiwe/avito-pr-service/pkg/postgres"
	"github.com/google/uuid"
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

func (r *Repository) Create(ctx context.Context, ID, name string, teamID uuid.UUID, isActive bool) (entity.User, error) {
	logrus.Infof("UserRepository.Create: creating user with name %s", name)

	query, args, _ := r.Builder.Insert("users").
		Columns("id", "name", "team_id", "is_active").
		Values(ID, name, teamID, isActive).
		Suffix("RETURNING created_at").
		ToSql()

	user := entity.User{
		ID:       ID,
		Name:     name,
		TeamID:   teamID,
		IsActive: isActive,
	}

	err := r.GetTxManager(ctx).QueryRow(ctx, query, args...).Scan(
		&user.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				logrus.Warnf("UserRepository.Create: user already exists: %s", name)
				return entity.User{}, repository.ErrUserAlreadyExists
			}
		}
		logrus.Errorf("UserRepository.Create: failed to create user: %v", err)
		return entity.User{}, err
	}

	logrus.Infof("UserRepository.Create: user created with ID %s", user.ID)
	return user, nil
}

func (r *Repository) GetByID(ctx context.Context, ID string) (entity.User, error) {
	logrus.Infof("UserRepository.GetByID: getting user by ID %s", ID)

	query, args, _ := r.Builder.Select("id", "name", "is_active", "team_id", "created_at").
		From("users").
		Where("id = ?", ID).
		ToSql()

	var user entity.User

	err := r.GetTxManager(ctx).QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Name,
		&user.IsActive,
		&user.TeamID,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Warnf("UserRepository.GetByID: no user with ID %s", ID)
			return entity.User{}, repository.ErrUserNotFound
		}
		logrus.Errorf("UserRepository.GetByID: failed to get user by ID %s: %v", ID, err)
		return entity.User{}, err
	}

	logrus.Infof("UserRepository.GetByID: user found with ID %s", user.ID)
	return user, nil
}

func (r *Repository) GetByTeamID(ctx context.Context, teamID uuid.UUID) ([]entity.User, error) {
	logrus.Infof("UserRepository.GetByTeamID: getting users by team ID %s", teamID)

	query, args, _ := r.Builder.Select("id", "name", "is_active", "team_id", "created_at").
		From("users").
		Where("team_id = ?", teamID).
		ToSql()

	rows, err := r.GetTxManager(ctx).Query(ctx, query, args...)
	if err != nil {
		logrus.Errorf("UserRepository.GetByTeamID: failed to query users by team ID %s: %v", teamID, err)
		return nil, fmt.Errorf("UserRepository.GetByTeamID: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.IsActive,
			&user.TeamID,
			&user.CreatedAt,
		); err != nil {
			logrus.Errorf("UserRepository.GetByTeamID: failed to scan user row for team ID %s: %v", teamID, err)
			return nil, fmt.Errorf("UserRepository.GetByTeamID: %w", err)
		}
		users = append(users, user)
	}

	logrus.Infof("UserRepository.GetByTeamID: found %d users for team ID %s", len(users), teamID)
	return users, nil
}

func (r *Repository) SetTeamID(ctx context.Context, userID string, teamID uuid.UUID) error {
	logrus.Infof("UserRepository.SetTeamID: setting team ID %s for user ID %s", teamID, userID)

	query, args, _ := r.Builder.Update("users").
		Set("team_id", teamID).
		Where("id = ?", userID).
		ToSql()

	cmdTag, err := r.GetTxManager(ctx).Exec(ctx, query, args...)
	if err != nil {
		logrus.Errorf("UserRepository.SetTeamID: failed to set team ID %s for user ID %s: %v", teamID, userID, err)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		logrus.Warnf("UserRepository.SetTeamID: no user found with ID %s to set team ID %s", userID, teamID)
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *Repository) SetActiveStatus(ctx context.Context, userID string, isActive bool) error {
	logrus.Infof("UserRepository.SetActiveStatus: setting isActive=%t for user ID %s", isActive, userID)

	query, args, _ := r.Builder.Update("users").
		Set("is_active", isActive).
		Where("id = ?", userID).
		ToSql()

	cmdTag, err := r.GetTxManager(ctx).Exec(ctx, query, args...)
	if err != nil {
		logrus.Errorf("UserRepository.SetActiveStatus: failed to set isActive=%t for user ID %s: %v", isActive, userID, err)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		logrus.Warnf("UserRepository.SetActiveStatus: no user found with ID %s to set isActive=%t", userID, isActive)
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *Repository) GetRandomActiveTeammates(ctx context.Context, teamID uuid.UUID, excludeUserID string, limit int) ([]entity.User, error) {
	logrus.Infof("UserRepository.GetRandomActiveTeammates: getting up to %d random active teammates for user ID %s", limit, excludeUserID)

	query, args, _ := r.Builder.Select("id", "name", "team_id", "created_at").
		From("users").
		Where("team_id = ? AND is_active = TRUE AND id != ?", teamID, excludeUserID).
		OrderBy("RANDOM()").
		Limit(uint64(limit)).
		ToSql()

	rows, err := r.GetTxManager(ctx).Query(ctx, query, args...)
	if err != nil {
		logrus.Errorf("UserRepository.GetRandomActiveTeammates: failed to query random active teammates for user ID %s: %v", excludeUserID, err)
		return nil, fmt.Errorf("UserRepository.GetRandomActiveTeammates: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.TeamID,
			&user.CreatedAt,
		); err != nil {
			logrus.Errorf("UserRepository.GetRandomActiveTeammates: failed to scan user row for user ID %s: %v", excludeUserID, err)
			return nil, fmt.Errorf("UserRepository.GetRandomActiveTeammates: %w", err)
		}
		users = append(users, user)
	}
	logrus.Infof("UserRepository.GetRandomActiveTeammates: found %d random active teammates for user ID %s", len(users), excludeUserID)
	return users, nil
}
