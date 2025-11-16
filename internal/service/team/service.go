package team

import (
	"context"
	"errors"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	"github.com/4udiwe/avito-pr-service/pkg/transactor"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type Service struct {
	userRepo  UserRepo
	teamRepo  TeamRepo
	prRepo    PRRepo
	txManager transactor.Transactor
}

func New(userRepo UserRepo, teamRepo TeamRepo, prRepo PRRepo, txManager transactor.Transactor) *Service {
	return &Service{
		userRepo:  userRepo,
		teamRepo:  teamRepo,
		prRepo:    prRepo,
		txManager: txManager,
	}
}

func (s *Service) CreateTeamWithUsers(ctx context.Context, teamName string, users []entity.User) (entity.Team, error) {
	logrus.Infof("TeamService.CreateTeamWithUsers: creating team %s with %d users", teamName, len(users))

	var team entity.Team

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Create a team
		newTeam, err := s.teamRepo.Create(ctx, teamName)
		if err != nil {
			return err
		}

		team = newTeam

		// Create users
		team.Members, err = s.userRepo.CreateUsersBatch(ctx, users, team.ID)
		if err != nil {
			return err
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
		rowTeam, err := s.teamRepo.GetByName(ctx, teamName)
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

	teams, total, err := s.teamRepo.GetAll(ctx, limit, offset)
	if err != nil {
		logrus.Errorf("TeamService.GetAllTeams: failed to fetch teams %v", err)
		return nil, 0, ErrCannotFetchTeams
	}

	logrus.Infof("TeamService.GetAllTeams: fetched %d teams", len(teams))
	return teams, total, nil
}

// Deactivates all team members and reassigns them on open PRs with new random reviewers from other teams
func (s *Service) DeactivateTeamAndReassignPRs(ctx context.Context, teamName string) error {
	logrus.Infof("TeamService.DeactivateTeamAndReassignPRs: deactivating team %s and reassigning PRs", teamName)

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Deactivate team members
		users, err := s.teamRepo.DeactivateTeamMembers(ctx, teamName)
		if err != nil {
			return err
		}
		if len(users) == 0 {
			logrus.Infof("TeamService.DeactivateTeamAndReassignPRs: no active users found for team %s", teamName)
			return nil
		}

		// Get IDs of deactivated users
		deactivatedIDs := lo.Map(users, func(u entity.User, _ int) string { return u.ID })

		// For each deactivated fetch PRs, where he was a reviewer
		for _, userID := range deactivatedIDs {
			prs, err := s.prRepo.ListByReviewer(ctx, userID)
			if err != nil {
				return err
			}

			for _, pr := range prs {
				// For each PR get new random reviewers from other teams
				newReviewers, err := s.userRepo.GetRandomActiveUsers(
					ctx,
					1,                                      // Get only one new reviewer instead deactivated
					append(deactivatedIDs, pr.AuthorID)..., // Exclude deactivated users and author himself
				)
				if err != nil {
					return err
				}
				if len(newReviewers) > 1 {
					return ErrCannotFetchNewReviewer
				}

				// Reassign deactivated reviewer with new random
				if err := s.prRepo.ReassignReviewer(ctx, pr.ID, userID, newReviewers[0].ID); err != nil {
					return err
				}
			}
		}
		return err

	})

	if err != nil {
		return ErrCannotDeactivateTeam
	}

	logrus.Infof("TeamService.DeactivateTeamAndReassignPRs: completed for team %s", teamName)
	return nil
}
