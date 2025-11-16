package team_test

import (
	"context"
	"errors"
	"testing"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	mock_transactor "github.com/4udiwe/avito-pr-service/internal/mocks"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	"github.com/4udiwe/avito-pr-service/internal/service/team"
	"github.com/4udiwe/avito-pr-service/internal/service/team/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestService_CreateTeamWithUsers(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		setup func(
			u *mocks.MockUserRepo,
			tr *mocks.MockTeamRepo,
			pr *mocks.MockPRRepo,
			tx *mock_transactor.MockTransactor,
		)
		expectedErr error
	}{
		{
			name: "team already exists",
			setup: func(
				u *mocks.MockUserRepo,
				tr *mocks.MockTeamRepo,
				pr *mocks.MockPRRepo,
				tx *mock_transactor.MockTransactor,
			) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				tr.EXPECT().
					Create(gomock.Any(), "backend").
					Return(entity.Team{}, repository.ErrTeamAlreadyExists)
			},
			expectedErr: team.ErrTeamAlreadyExists,
		},

		{
			name: "user already exists",
			setup: func(
				u *mocks.MockUserRepo,
				tr *mocks.MockTeamRepo,
				pr *mocks.MockPRRepo,
				tx *mock_transactor.MockTransactor,
			) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				createdTeam := entity.Team{ID: uuid.New(), Name: "backend"}

				tr.EXPECT().
					Create(gomock.Any(), "backend").
					Return(createdTeam, nil)

				u.EXPECT().
					CreateUsersBatch(gomock.Any(), gomock.Any(), createdTeam.ID).
					Return([]entity.User{}, repository.ErrUserAlreadyExists)
			},
			expectedErr: team.ErrUserAlreadyExists,
		},

		{
			name: "success",
			setup: func(
				u *mocks.MockUserRepo,
				tr *mocks.MockTeamRepo,
				pr *mocks.MockPRRepo,
				tx *mock_transactor.MockTransactor,
			) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				createdTeam := entity.Team{ID: uuid.New(), Name: "backend"}

				users := make([]entity.User, 1)
				users = append(users, entity.User{ID: "1", Name: "John", Team: entity.Team{ID: createdTeam.ID}, IsActive: true})

				tr.EXPECT().
					Create(gomock.Any(), "backend").
					Return(createdTeam, nil)

				u.EXPECT().
					CreateUsersBatch(gomock.Any(), gomock.Any(), createdTeam.ID).
					Return(users, nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			u := mocks.NewMockUserRepo(ctrl)
			tr := mocks.NewMockTeamRepo(ctrl)
			pr := mocks.NewMockPRRepo(ctrl)
			tx := mock_transactor.NewMockTransactor(ctrl)

			tt.setup(u, tr, pr, tx)

			svc := team.New(u, tr, pr, tx)

			_, err := svc.CreateTeamWithUsers(ctx, "backend", []entity.User{
				{ID: "1", Name: "John", IsActive: true},
			})

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected: %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestService_GetTeamWithMembers(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		setup func(
			u *mocks.MockUserRepo,
			tr *mocks.MockTeamRepo,
			pr *mocks.MockPRRepo,
			tx *mock_transactor.MockTransactor,
		)
		expectedErr error
	}{
		{
			name: "team not found",
			setup: func(
				u *mocks.MockUserRepo,
				tr *mocks.MockTeamRepo,
				pr *mocks.MockPRRepo,
				tx *mock_transactor.MockTransactor,
			) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				tr.EXPECT().
					GetByName(gomock.Any(), "backend").
					Return(entity.Team{}, repository.ErrTeamNotFound)
			},
			expectedErr: team.ErrTeamNotFound,
		},

		{
			name: "success",
			setup: func(
				u *mocks.MockUserRepo,
				tr *mocks.MockTeamRepo,
				pr *mocks.MockPRRepo,
				tx *mock_transactor.MockTransactor,
			) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				createdTeam := entity.Team{ID: uuid.New(), Name: "backend"}

				tr.EXPECT().
					GetByName(gomock.Any(), "backend").
					Return(createdTeam, nil)

				u.EXPECT().
					GetByTeamID(gomock.Any(), createdTeam.ID).
					Return([]entity.User{
						{ID: "1", Name: "John"},
						{ID: "2", Name: "Jane"},
					}, nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			u := mocks.NewMockUserRepo(ctrl)
			tr := mocks.NewMockTeamRepo(ctrl)
			pr := mocks.NewMockPRRepo(ctrl)
			tx := mock_transactor.NewMockTransactor(ctrl)

			tt.setup(u, tr, pr, tx)

			svc := team.New(u, tr, pr, tx)

			_, err := svc.GetTeamWithMembers(ctx, "backend")

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected: %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestService_GetAllTeams(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		setup func(
			u *mocks.MockUserRepo,
			tr *mocks.MockTeamRepo,
			pr *mocks.MockPRRepo,
		)
		expectedErr error
	}{
		{
			name: "repo error",
			setup: func(
				u *mocks.MockUserRepo,
				tr *mocks.MockTeamRepo,
				pr *mocks.MockPRRepo,
			) {
				tr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return(nil, 0, errors.New("db err"))
			},
			expectedErr: team.ErrCannotFetchTeams,
		},
		{
			name: "success",
			setup: func(
				u *mocks.MockUserRepo,
				tr *mocks.MockTeamRepo,
				pr *mocks.MockPRRepo,
			) {
				tr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Team{
						{ID: uuid.New(), Name: "backend"},
						{ID: uuid.New(), Name: "ml"},
					}, 2, nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			u := mocks.NewMockUserRepo(ctrl)
			tr := mocks.NewMockTeamRepo(ctrl)
			pr := mocks.NewMockPRRepo(ctrl)

			svc := team.New(u, tr, pr, nil)

			tt.setup(u, tr, pr)

			_, _, err := svc.GetAllTeams(ctx, 1, 10)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestService_DeactivateTeamAndReassignPRs(t *testing.T) {
	ctx := context.Background()

	deactivateResult := []entity.User{
		{ID: "u1", Name: "John"},
		{ID: "u2", Name: "Jane"},
	}

	prList := []entity.PullRequest{
		{
			ID:        "pr1",
			AuthorID:  "a1",
			Reviewers: []string{"u1"},
		},
	}

	tests := []struct {
		name  string
		setup func(
			u *mocks.MockUserRepo,
			tr *mocks.MockTeamRepo,
			pr *mocks.MockPRRepo,
			tx *mock_transactor.MockTransactor,
		)
		expectedErr error
	}{
		{
			name: "deactivation error",
			setup: func(u *mocks.MockUserRepo, tr *mocks.MockTeamRepo, pr *mocks.MockPRRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				tr.EXPECT().
					DeactivateTeamMembers(gomock.Any(), "backend").
					Return(nil, errors.New("db"))
			},
			expectedErr: team.ErrCannotDeactivateTeam,
		},

		{
			name: "no users → no PR actions",
			setup: func(u *mocks.MockUserRepo, tr *mocks.MockTeamRepo, pr *mocks.MockPRRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				tr.EXPECT().
					DeactivateTeamMembers(gomock.Any(), "backend").
					Return([]entity.User{}, nil)
			},
			expectedErr: nil,
		},

		{
			name: "ListByReviewer error",
			setup: func(u *mocks.MockUserRepo, tr *mocks.MockTeamRepo, pr *mocks.MockPRRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				tr.EXPECT().
					DeactivateTeamMembers(gomock.Any(), "backend").
					Return(deactivateResult, nil)

				pr.EXPECT().
					ListByReviewer(gomock.Any(), "u1").
					Return(nil, errors.New("db"))
			},
			expectedErr: team.ErrCannotDeactivateTeam,
		},

		{
			name: "GetRandomActiveUsers error",
			setup: func(u *mocks.MockUserRepo, tr *mocks.MockTeamRepo, pr *mocks.MockPRRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				tr.EXPECT().
					DeactivateTeamMembers(gomock.Any(), "backend").
					Return(deactivateResult, nil)

				pr.EXPECT().
					ListByReviewer(gomock.Any(), "u1").
					Return(prList, nil)

				u.EXPECT().
					GetRandomActiveUsers(gomock.Any(), 1, gomock.Any()).
					Return(nil, errors.New("db"))
			},
			expectedErr: team.ErrCannotDeactivateTeam,
		},

		{
			name: "GetRandomActiveUsers returned >1 user",
			setup: func(u *mocks.MockUserRepo, tr *mocks.MockTeamRepo, pr *mocks.MockPRRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				tr.EXPECT().
					DeactivateTeamMembers(gomock.Any(), "backend").
					Return(deactivateResult, nil)

				pr.EXPECT().
					ListByReviewer(gomock.Any(), "u1").
					Return(prList, nil)

				u.EXPECT().
					GetRandomActiveUsers(gomock.Any(), 1, gomock.Any()).
					Return([]entity.User{
						{ID: "x1"},
						{ID: "x2"},
					}, nil)
			},
			expectedErr: team.ErrCannotDeactivateTeam,
		},

		{
			name: "ReassignReviewer error",
			setup: func(u *mocks.MockUserRepo, tr *mocks.MockTeamRepo, pr *mocks.MockPRRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				tr.EXPECT().
					DeactivateTeamMembers(gomock.Any(), "backend").
					Return(deactivateResult, nil)

				pr.EXPECT().
					ListByReviewer(gomock.Any(), "u1").
					Return(prList, nil)

				u.EXPECT().
					GetRandomActiveUsers(gomock.Any(), 1, gomock.Any()).
					Return([]entity.User{{ID: "newrev"}}, nil)

				pr.EXPECT().
					ReassignReviewer(gomock.Any(), "pr1", "u1", "newrev").
					Return(errors.New("db"))
			},
			expectedErr: team.ErrCannotDeactivateTeam,
		},

		{
			name: "success",
			setup: func(u *mocks.MockUserRepo, tr *mocks.MockTeamRepo, pr *mocks.MockPRRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				tr.EXPECT().
					DeactivateTeamMembers(gomock.Any(), "backend").
					Return(deactivateResult, nil)

				/* for u1 */
				pr.EXPECT().
					ListByReviewer(gomock.Any(), "u1").
					Return(prList, nil)

				u.EXPECT().
					GetRandomActiveUsers(gomock.Any(), 1, gomock.Any()).
					Return([]entity.User{{ID: "r1"}}, nil)

				pr.EXPECT().
					ReassignReviewer(gomock.Any(), "pr1", "u1", "r1").
					Return(nil)

				/* for u2 — no PRs */
				pr.EXPECT().
					ListByReviewer(gomock.Any(), "u2").
					Return([]entity.PullRequest{}, nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			u := mocks.NewMockUserRepo(ctrl)
			tr := mocks.NewMockTeamRepo(ctrl)
			pr := mocks.NewMockPRRepo(ctrl)
			tx := mock_transactor.NewMockTransactor(ctrl)

			svc := team.New(u, tr, pr, tx)

			tt.setup(u, tr, pr, tx)

			err := svc.DeactivateTeamAndReassignPRs(ctx, "backend")

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
