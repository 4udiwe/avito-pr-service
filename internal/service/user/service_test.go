package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	mock_transactor "github.com/4udiwe/avito-pr-service/internal/mocks"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	service "github.com/4udiwe/avito-pr-service/internal/service/user"
	"github.com/4udiwe/avito-pr-service/internal/service/user/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSetUserStatus(t *testing.T) {
	var (
		ctx          = context.Background()
		userID       = "user123"
		arbitraryErr = errors.New("arbitrary error")
	)

	mockUser := entity.User{
		ID:       userID,
		Name:     "John",
		IsActive: true,
	}

	type MockBehavior func(
		u *mocks.MockUserRepo,
	)

	for _, tc := range []struct {
		name         string
		userID       string
		isActive     bool
		mockBehavior MockBehavior
		want         entity.User
		wantErr      error
	}{
		{
			name:     "success",
			userID:   userID,
			isActive: false,
			mockBehavior: func(u *mocks.MockUserRepo) {
				u.EXPECT().SetActiveStatus(ctx, userID, false).Return(nil).Times(1)
				expectedUser := mockUser
				expectedUser.IsActive = false
				u.EXPECT().GetByID(ctx, userID).Return(expectedUser, nil).Times(1)
			},
			want: func() entity.User {
				u := mockUser
				u.IsActive = false
				return u
			}(),
			wantErr: nil,
		},
		{
			name:     "user not found on status update",
			userID:   userID,
			isActive: false,
			mockBehavior: func(u *mocks.MockUserRepo) {
				u.EXPECT().SetActiveStatus(ctx, userID, false).Return(repository.ErrUserNotFound).Times(1)
			},
			want:    entity.User{},
			wantErr: service.ErrUserNotFound,
		},
		{
			name:     "internal error on status update",
			userID:   userID,
			isActive: true,
			mockBehavior: func(u *mocks.MockUserRepo) {
				u.EXPECT().SetActiveStatus(ctx, userID, true).Return(arbitraryErr).Times(1)
			},
			want:    entity.User{},
			wantErr: service.ErrCannotSetUserStatus,
		},
		{
			name:     "user not found on GetByID",
			userID:   userID,
			isActive: true,
			mockBehavior: func(u *mocks.MockUserRepo) {
				u.EXPECT().SetActiveStatus(ctx, userID, true).Return(nil).Times(1)
				u.EXPECT().GetByID(ctx, userID).Return(entity.User{}, repository.ErrUserNotFound).Times(1)
			},
			want:    entity.User{},
			wantErr: service.ErrUserNotFound,
		},
		{
			name:     "internal error on GetByID",
			userID:   userID,
			isActive: true,
			mockBehavior: func(u *mocks.MockUserRepo) {
				u.EXPECT().SetActiveStatus(ctx, userID, true).Return(nil).Times(1)
				u.EXPECT().GetByID(ctx, userID).Return(entity.User{}, arbitraryErr).Times(1)
			},
			want:    entity.User{},
			wantErr: service.ErrCannotSetUserStatus,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			MockUserRepo := mocks.NewMockUserRepo(ctrl)
			MockPRRepo := mocks.NewMockPullReqeustRepo(ctrl)
			MockTransactor := mock_transactor.NewMockTransactor(ctrl)

			tc.mockBehavior(MockUserRepo)

			s := service.New(MockUserRepo, MockPRRepo, MockTransactor)

			out, err := s.SetUserStatus(ctx, tc.userID, tc.isActive)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, out)
		})
	}
}

func TestGetUserReviews(t *testing.T) {
	ctx := context.Background()
	userID := "123"

	mockUser := entity.User{
		ID:   userID,
		Name: "Test",
	}

	mockReviewers := make([]string, 2)
	mockReviewers = append(mockReviewers, "u1", "u2")

	mockPRs := []entity.PullRequest{
		{ID: "1", AuthorID: "a1", Reviewers: mockReviewers},
		{ID: "2", AuthorID: "a2", Reviewers: mockReviewers},
	}

	arbitraryErr := errors.New("unexpected failure")

	tests := []struct {
		name  string
		setup func(
			u *mocks.MockUserRepo,
			p *mocks.MockPullReqeustRepo,
			tx *mock_transactor.MockTransactor,
		)
		expectedPRs []entity.PullRequest
		expectedErr error
	}{
		{
			name: "success",
			setup: func(
				u *mocks.MockUserRepo,
				p *mocks.MockPullReqeustRepo,
				tx *mock_transactor.MockTransactor,
			) {
				tx.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				u.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(mockUser, nil)

				p.EXPECT().
					ListByReviewer(gomock.Any(), userID).
					Return(mockPRs, nil)
			},
			expectedPRs: mockPRs,
			expectedErr: nil,
		},
		{
			name: "user not found",
			setup: func(
				u *mocks.MockUserRepo,
				p *mocks.MockPullReqeustRepo,
				tx *mock_transactor.MockTransactor,
			) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				u.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(entity.User{}, repository.ErrUserNotFound)

			},
			expectedPRs: nil,
			expectedErr: service.ErrUserNotFound,
		},
		{
			name: "cannot fetch PRs",
			setup: func(
				u *mocks.MockUserRepo,
				p *mocks.MockPullReqeustRepo,
				tx *mock_transactor.MockTransactor,
			) {
				tx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				u.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(mockUser, nil)

				p.EXPECT().
					ListByReviewer(gomock.Any(), userID).
					Return(nil, arbitraryErr)
			},
			expectedPRs: nil,
			expectedErr: service.ErrCannotGetUserReviews,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mocks.NewMockUserRepo(ctrl)
			mockPRRepo := mocks.NewMockPullReqeustRepo(ctrl)
			mockTx := mock_transactor.NewMockTransactor(ctrl)

			tt.setup(mockUserRepo, mockPRRepo, mockTx)

			s := service.New(mockUserRepo, mockPRRepo, mockTx)

			out, err := s.GetUserReviews(ctx, userID)

			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.expectedPRs, out)
		})
	}
}
