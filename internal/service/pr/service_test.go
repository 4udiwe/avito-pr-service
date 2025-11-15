package pr_test

import (
	"context"
	"errors"
	"testing"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	mock_transactor "github.com/4udiwe/avito-pr-service/internal/mocks"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	service "github.com/4udiwe/avito-pr-service/internal/service/pr"
	"github.com/4udiwe/avito-pr-service/internal/service/pr/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestService_CreatePR(t *testing.T) {
	ctx := context.Background()

	author := entity.User{
		ID:   "author1",
		Name: "Alice",
		Team: entity.Team{ID: uuid.New()},
	}

	tests := []struct {
		name  string
		setup func(
			pr *mocks.MockPRRepo,
			u *mocks.MockUserRepo,
			tx *mock_transactor.MockTransactor,
		)
		expectedErr error
	}{
		{
			name: "author not found",
			setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				u.EXPECT().
					GetByID(gomock.Any(), "author1").
					Return(entity.User{}, repository.ErrUserNotFound)
			},
			expectedErr: service.ErrAuthorNotFound,
		},

		{
			name: "GetRandomActiveTeammates error",
			setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				u.EXPECT().GetByID(gomock.Any(), "author1").
					Return(author, nil)

				u.EXPECT().
					GetRandomActiveTeammates(gomock.Any(), author.Team.ID, 2, "author1").
					Return(nil, errors.New("db"))
			},
			expectedErr: service.ErrCannotCreatePR,
		},

		{
			name: "PRRepo.Create error",
			setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				u.EXPECT().GetByID(gomock.Any(), "author1").Return(author, nil)
				u.EXPECT().GetRandomActiveTeammates(gomock.Any(), author.Team.ID, 2, "author1").
					Return([]entity.User{{ID: "r1"}}, nil)

				pr.EXPECT().
					Create(gomock.Any(), "pr1", "title", "author1", "OPEN", true).
					Return(entity.PullRequest{}, errors.New("db"))
			},
			expectedErr: service.ErrCannotCreatePR,
		},

		{
			name: "AssignReviewers error",
			setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				u.EXPECT().GetByID(gomock.Any(), "author1").Return(author, nil)
				u.EXPECT().GetRandomActiveTeammates(gomock.Any(), author.Team.ID, 2, "author1").
					Return([]entity.User{{ID: "r1"}}, nil)

				pr.EXPECT().
					Create(gomock.Any(), "pr1", "title", "author1", "OPEN", true).
					Return(entity.PullRequest{ID: "pr1"}, nil)

				pr.EXPECT().
					AssignReviewers(gomock.Any(), "pr1", []string{"r1"}).
					Return(errors.New("db"))
			},
			expectedErr: service.ErrCannotCreatePR,
		},

		{
			name: "success (needMoreReviewers = true)",
			setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				u.EXPECT().GetByID(gomock.Any(), "author1").Return(author, nil)
				u.EXPECT().GetRandomActiveTeammates(gomock.Any(), author.Team.ID, 2, "author1").
					Return([]entity.User{{ID: "r1"}}, nil)

				pr.EXPECT().
					Create(gomock.Any(), "pr1", "title", "author1", "OPEN", true).
					Return(entity.PullRequest{ID: "pr1"}, nil)

				pr.EXPECT().
					AssignReviewers(gomock.Any(), "pr1", []string{"r1"}).
					Return(nil)
			},
			expectedErr: nil,
		},

		{
			name: "success (needMoreReviewers = false)",
			setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				u.EXPECT().GetByID(gomock.Any(), "author1").Return(author, nil)
				u.EXPECT().GetRandomActiveTeammates(gomock.Any(), author.Team.ID, 2, "author1").
					Return([]entity.User{{ID: "r1"}, {ID: "r2"}}, nil)

				pr.EXPECT().
					Create(gomock.Any(), "pr1", "title", "author1", "OPEN", false).
					Return(entity.PullRequest{ID: "pr1"}, nil)

				pr.EXPECT().
					AssignReviewers(gomock.Any(), "pr1", []string{"r1", "r2"}).
					Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			pr := mocks.NewMockPRRepo(ctrl)
			u := mocks.NewMockUserRepo(ctrl)
			tx := mock_transactor.NewMockTransactor(ctrl)

			svc := service.New(pr, u, tx)

			tt.setup(pr, u, tx)

			_, err := svc.CreatePR(ctx, "pr1", "title", "author1")
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestService_GetAllPRs(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		setup       func(pr *mocks.MockPRRepo)
		expectedErr error
	}{
		{
			name: "repo error",
			setup: func(pr *mocks.MockPRRepo) {
				pr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return(nil, 0, errors.New("db"))
			},
			expectedErr: service.ErrCannotFetchPRs,
		},

		{
			name: "success",
			setup: func(pr *mocks.MockPRRepo) {
				pr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.PullRequest{
						{ID: "pr1", Title: "one"},
						{ID: "pr2", Title: "two"},
					}, 2, nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			prRepo := mocks.NewMockPRRepo(ctrl)
			u := mocks.NewMockUserRepo(ctrl)

			svc := service.New(prRepo, u, nil)

			tt.setup(prRepo)

			_, _, err := svc.GetAllPRs(ctx, 1, 10)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestService_ReassignReviewer(t *testing.T) {
	ctx := context.Background()
	prID := "pr1"
	oldReviewerID := "rev1"

	author := entity.User{ID: "author1", Team: entity.Team{ID: uuid.New()}}
	oldReviewer := entity.User{ID: oldReviewerID, Team: author.Team}
	openStatus := entity.Status{ID: 1, Name: entity.StatusOPEN}

	tests := []struct {
		name        string
		setup       func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor)
		expectedErr error
	}{
		{
			name: "PR not found",
			setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)

				pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{}, repository.ErrPRNotFound)
			},
			expectedErr: service.ErrPRNotFound,
		},
		{
			name: "old reviewer not found",
			setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)

				pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{ID: prID, Status: openStatus, AuthorID: author.ID}, nil)
				u.EXPECT().GetByID(gomock.Any(), oldReviewerID).Return(entity.User{}, repository.ErrUserNotFound)
			},
			expectedErr: service.ErrReviewerNotFound,
		},
		{
			name: "no more reviewers to reassign",
			setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)

				PR := entity.PullRequest{ID: prID, Status: openStatus, AuthorID: author.ID}
				emptyUsers := []entity.User{}

				pr.EXPECT().GetByID(gomock.Any(), prID).Return(PR, nil)
				u.EXPECT().GetByID(gomock.Any(), oldReviewerID).Return(oldReviewer, nil)
				u.EXPECT().GetRandomActiveTeammates(gomock.Any(), oldReviewer.Team.ID, 1, PR.AuthorID, oldReviewerID).Return(emptyUsers, nil)
			},
			expectedErr: service.ErrNoMoreReviewersToReassign,
		},
		{
			name: "ReassignReviewer fails",
			setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)

				pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{ID: prID, Status: openStatus, AuthorID: author.ID}, nil)
				u.EXPECT().GetByID(gomock.Any(), oldReviewerID).Return(oldReviewer, nil)
				u.EXPECT().GetRandomActiveTeammates(gomock.Any(), oldReviewer.Team.ID, 1, author.ID, oldReviewerID).
					Return([]entity.User{{ID: "newRev"}}, nil)
				pr.EXPECT().ReassignReviewer(gomock.Any(), prID, oldReviewerID, "newRev").Return(errors.New("db"))
			},
			expectedErr: service.ErrCannotAssignReviewer,
		},
		{
			name: "success",
			setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
				tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)

				pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{ID: prID, Status: openStatus, AuthorID: author.ID}, nil)
				u.EXPECT().GetByID(gomock.Any(), oldReviewerID).Return(oldReviewer, nil)
				u.EXPECT().GetRandomActiveTeammates(gomock.Any(), oldReviewer.Team.ID, 1, author.ID, oldReviewerID).
					Return([]entity.User{{ID: "newRev"}}, nil)
				pr.EXPECT().ReassignReviewer(gomock.Any(), prID, oldReviewerID, "newRev").Return(nil)
				pr.EXPECT().GetReviewersByPR(gomock.Any(), prID).Return([]entity.PRReviewer{{PRID: prID, ReviewerID: "newRev"}}, nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			prRepo := mocks.NewMockPRRepo(ctrl)
			uRepo := mocks.NewMockUserRepo(ctrl)
			tx := mock_transactor.NewMockTransactor(ctrl)

			svc := service.New(prRepo, uRepo, tx)

			tt.setup(prRepo, uRepo, tx)

			_, _, err := svc.ReassignReviewer(ctx, prID, oldReviewerID)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestService_MergePR(t *testing.T) {
    ctx := context.Background()
    prID := "pr1"

    openStatus := entity.Status{ID: 1, Name: entity.StatusOPEN}
    mergedStatus := entity.Status{ID: 2, Name: entity.StatusMERGED}

    tests := []struct {
        name        string
        setup       func(pr *mocks.MockPRRepo)
        expectedErr error
    }{
        {
            name: "PR not found",
            setup: func(pr *mocks.MockPRRepo) {
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{}, repository.ErrPRNotFound)
            },
            expectedErr: service.ErrPRNotFound,
        },
        {
            name: "PR already merged",
            setup: func(pr *mocks.MockPRRepo) {
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{ID: prID, Status: mergedStatus}, nil)
            },
            expectedErr: nil,
        },
        {
            name: "GetPRStatuses fails",
            setup: func(pr *mocks.MockPRRepo) {
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{ID: prID, Status: openStatus}, nil)
                pr.EXPECT().GetPRStatuses(gomock.Any()).Return(nil, errors.New("db"))
            },
            expectedErr: service.ErrCannotFetchStatus,
        },
        {
            name: "UpdateStatus fails",
            setup: func(pr *mocks.MockPRRepo) {
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{ID: prID, Status: openStatus}, nil)
                pr.EXPECT().GetPRStatuses(gomock.Any()).Return([]entity.Status{mergedStatus}, nil)
                pr.EXPECT().UpdateStatus(gomock.Any(), prID, mergedStatus.ID, gomock.Any()).Return(errors.New("db"))
            },
            expectedErr: service.ErrCannotCreatePR,
        },
        {
            name: "success",
            setup: func(pr *mocks.MockPRRepo) {
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{ID: prID, Status: openStatus}, nil)
                pr.EXPECT().GetPRStatuses(gomock.Any()).Return([]entity.Status{mergedStatus}, nil)
                pr.EXPECT().UpdateStatus(gomock.Any(), prID, mergedStatus.ID, gomock.Any()).Return(nil)
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{ID: prID, Status: mergedStatus}, nil)
            },
            expectedErr: nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            prRepo := mocks.NewMockPRRepo(ctrl)
            uRepo := mocks.NewMockUserRepo(ctrl)

            svc := service.New(prRepo, uRepo, nil)

            tt.setup(prRepo)

            _, err := svc.MergePR(ctx, prID)
            if !errors.Is(err, tt.expectedErr) {
                t.Fatalf("expected %v, got %v", tt.expectedErr, err)
            }
        })
    }
}

func TestService_AssignReviewer(t *testing.T) {
    ctx := context.Background()
    prID := "pr1"
    newReviewerID := "rev2"

    openPR := entity.PullRequest{
        ID:        prID,
        NeedMoreReviewers: true,
        Reviewers: []string{"rev1"},
    }

    tests := []struct {
        name        string
        setup       func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor)
        expectedErr error
    }{
        {
            name: "PR not found",
            setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
                tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
                    func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
                )
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{}, repository.ErrPRNotFound)
            },
            expectedErr: service.ErrPRNotFound,
        },
        {
            name: "PR already has 2 reviewers",
            setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
                tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
                    func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
                )
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(entity.PullRequest{ID: prID, NeedMoreReviewers: false}, nil)
            },
            expectedErr: service.ErrPRAlreadyHas2Reviewers,
        },
        {
            name: "AssignReviewer fails ErrReviewerNotFound",
            setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
                tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
                    func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
                )
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(openPR, nil)
                pr.EXPECT().AssignReviewer(gomock.Any(), prID, newReviewerID).Return(repository.ErrReviewerNotFound)
            },
            expectedErr: service.ErrReviewerNotFound,
        },
        {
            name: "AssignReviewer fails ErrReviewerAlreadyAssigned",
            setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
                tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
                    func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
                )
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(openPR, nil)
                pr.EXPECT().AssignReviewer(gomock.Any(), prID, newReviewerID).Return(repository.ErrReviewerAlreadyAssigned)
            },
            expectedErr: service.ErrReviewerAlreadyAssigned,
        },
        {
            name: "UpdateNeedMoreReviewers fails",
            setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
                tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
                    func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
                )
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(openPR, nil)
                pr.EXPECT().AssignReviewer(gomock.Any(), prID, newReviewerID).Return(nil)
                pr.EXPECT().UpdateNeedMoreReviewers(gomock.Any(), prID).Return(errors.New("db"))
            },
            expectedErr: service.ErrCannotAssignReviewer,
        },
        {
            name: "success",
            setup: func(pr *mocks.MockPRRepo, u *mocks.MockUserRepo, tx *mock_transactor.MockTransactor) {
                tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
                    func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
                )
                pr.EXPECT().GetByID(gomock.Any(), prID).Return(openPR, nil)
                pr.EXPECT().AssignReviewer(gomock.Any(), prID, newReviewerID).Return(nil)
                pr.EXPECT().UpdateNeedMoreReviewers(gomock.Any(), prID).Return(nil)
                pr.EXPECT().GetReviewersByPR(gomock.Any(), prID).Return([]entity.PRReviewer{{PRID: prID, ReviewerID: "rev1"}, {PRID: prID, ReviewerID: newReviewerID}}, nil)
            },
            expectedErr: nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            prRepo := mocks.NewMockPRRepo(ctrl)
            uRepo := mocks.NewMockUserRepo(ctrl)
            tx := mock_transactor.NewMockTransactor(ctrl)

            svc := service.New(prRepo, uRepo, tx)

            tt.setup(prRepo, uRepo, tx)

            _, err := svc.AssignReviewer(ctx, prID, newReviewerID)
            if !errors.Is(err, tt.expectedErr) {
                t.Fatalf("expected %v, got %v", tt.expectedErr, err)
            }
        })
    }
}
