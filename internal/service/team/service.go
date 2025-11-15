package team

import (
	"context"
	"errors"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	"github.com/4udiwe/avito-pr-service/pkg/transactor"
	"github.com/sirupsen/logrus"
)

type Service struct {
	userRepo  UserRepo
	TeamRepo  TeamRepo
	txManager transactor.Transactor
}

func New(userRepo UserRepo, TeamRepo TeamRepo, txManager transactor.Transactor) *Service {
	return &Service{
		userRepo:  userRepo,
		TeamRepo:  TeamRepo,
		txManager: txManager,
	}
}

func (s *Service) CreateTeamWithUsers(ctx context.Context, teamName string, users []entity.User) (entity.Team, error) {
	logrus.Infof("TeamService.CreateTeamWithUsers: creating team %s with %d users", teamName, len(users))

	var team entity.Team

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Create a team
		newTeam, err := s.TeamRepo.Create(ctx, teamName)
		if err != nil {
			return err
		}

		team = newTeam

		// Create users
		for _, u := range users {
			newUser, err := s.userRepo.Create(ctx, u.ID, u.Name, team.ID, u.IsActive)
			if err != nil {
				return err
			}

			team.Members = append(team.Members, newUser)
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, repository.ErrTeamAlreadyExists) {
			return entity.Team{}, ErrTeamAlreadyExists
		}
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return entity.Team{}, ErrUserAlreadyExists
		}
		logrus.Errorf("TeamService.CreateTeamWithUsers: failed to create team %s: %v", teamName, err)
		return entity.Team{}, ErrCannotCreateTeam
	}

	return team, nil
}

func (s *Service) GetTeamWithMembers(ctx context.Context, teamName string) (entity.Team, error) {
	logrus.Infof("TeamService.GetTeamWithMembers: getting team %s with members", teamName)

	var team entity.Team

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Get a team
		rowTeam, err := s.TeamRepo.GetByName(ctx, teamName)
		if err != nil {
			return err
		}

		team = rowTeam

		// Get members of team
		users, err := s.userRepo.GetByTeamID(ctx, rowTeam.ID)
		if err != nil {
			return err
		}

		team.Members = users
		return nil
	})

	if err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			logrus.Warnf("TeamService.GetTeamWithMembers: team %s not found", teamName)
			return entity.Team{}, ErrTeamNotFound
		}
		logrus.Errorf("TeamService.GetTeamWithMembers: failed to get team %s: %v", teamName, err)
		return entity.Team{}, ErrCannotFetchTeam
	}

	return team, nil
}

func (s *Service) GetAllTeams(ctx context.Context, page, pageSize int) ([]entity.Team, int, error) {
	logrus.Info("TeamService.GetAllTeams: fetching all teams")

	limit := pageSize
	offset := (page - 1) * pageSize

	teams, total, err := s.TeamRepo.GetAll(ctx, limit, offset)
	if err != nil {
		logrus.Errorf("TeamService.GetAllTeams: failed to fetch teams %v", err)
		return nil, 0, ErrCannotFetchTeams
	}

	logrus.Infof("TeamService.GetAllTeams: fetched %d teams", len(teams))
	return teams, total, nil
}
