package service_test

import (
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/service"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock user verif repo
type MockUserVerificationRepo struct {
	mock.Mock
}

// func find by token
func (m *MockUserVerificationRepo) FindByToken(ctx context.Context, token string) (*entity.UserVerification, error) {
	args := m.Called(ctx, token)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.UserVerification), args.Error(1)
	}
	return nil, args.Error(1)
}

// func upsert
func (m *MockUserVerificationRepo) Upsert(ctx context.Context, data *entity.UserVerification) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

// func by email
func (m *MockUserVerificationRepo) FindByEmail(ctx context.Context, email string) (*entity.UserVerification, error) {
	args := m.Called(ctx, email)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.UserVerification), args.Error(1)
	}
	return nil, args.Error(1)
}

// mock user repo
type MockUserRepo struct {
	mock.Mock
}

// func find by id
func (m *MockUserRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// func verif user
func (m *MockUserRepo) VerifUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepo) FindByEmailUsername(ctx context.Context, value string) (*entity.User, error) {
	return nil, nil
}
func (m *MockUserRepo) Create(ctx context.Context, user *entity.User) error { return nil }
func (m *MockUserRepo) UpdateRefreshToken(ctx context.Context, id string, token string) error {
	return nil
}
func (m *MockUserRepo) FindByRefreshToken(ctx context.Context, token string) (*entity.User, error) {
	return nil, nil
}

// test verify user
func TestVerifyUser(t *testing.T) {
	// scenario: Token Valid and user verified successfully
	t.Run("Success Verification", func(t *testing.T) {
		ctx := context.Background()
		validToken := "TOKEN-123"
		userID := "USER-123"

		mockVerifData := &entity.UserVerification{
			UserID:    userID,
			Token:     validToken,
			ExpiredAt: time.Now().Add(1 * time.Hour),
		}

		mockUserData := &entity.User{
			IsVerified: false,
		}
		mockUserData.ID = userID

		// setup mocking
		mockAuthRepo := new(MockUserVerificationRepo)
		mockUserRepo := new(MockUserRepo)

		mockAuthRepo.On("FindByToken", ctx, validToken).Return(mockVerifData, nil)
		mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil)
		mockUserRepo.On("VerifUser", ctx, userID).Return(nil)

		serviceAuth := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil)
		err := serviceAuth.VerifyUser(ctx, validToken)

		assert.NoError(t, err)
		mockAuthRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	// failed scenario: token expired
	t.Run("Failed - Token Expired", func(t *testing.T) {
		ctx := context.Background()
		expiredToken := "TOKEN-EXPIRED"

		mockVerifData := &entity.UserVerification{
			UserID:    "USER-123",
			Token:     expiredToken,
			ExpiredAt: time.Now().Add(-1 * time.Hour),
		}

		mockAuthRepo := new(MockUserVerificationRepo)
		mockUserRepo := new(MockUserRepo)

		mockAuthRepo.On("FindByToken", ctx, expiredToken).Return(mockVerifData, nil)

		serviceAuth := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil)
		err := serviceAuth.VerifyUser(ctx, expiredToken)

		assert.ErrorIs(t, err, customerrors.ErrTokenExpired)
		mockAuthRepo.AssertExpectations(t)
	})

	// failed scenario: token not found
	t.Run("Failed - Token Not Found", func(t *testing.T) {
		ctx := context.Background()
		notFoundToken := "TOKEN-UNAVAIL"

		mockAuthRepo := new(MockUserVerificationRepo)
		MockUserRepo := new(MockUserRepo)

		expectedError := customerrors.ErrDataNotFound

		mockAuthRepo.On("FindByToken", ctx, notFoundToken).Return(nil, expectedError)

		serviceAuth := service.NewAuthService(MockUserRepo, nil, mockAuthRepo, nil, nil, nil)

		err := serviceAuth.VerifyUser(ctx, notFoundToken)

		assert.ErrorIs(t, err, expectedError)

		mockAuthRepo.AssertExpectations(t)
		MockUserRepo.AssertNotCalled(t, "FindByID")
	})

	// failed scenario: user not found
	t.Run("Failed - User Not Found", func(t *testing.T) {
		ctx := context.Background()
		validToken := "TOKEN-123"
		userID := "USER-123"

		mockVerifData := &entity.UserVerification{
			UserID: userID,
			Token: validToken,
			ExpiredAt: time.Now().Add(1 * time.Hour),
		}

		mockAuthRepo := new(MockUserVerificationRepo)
		mockUserRepo := new(MockUserRepo)

		mockAuthRepo.On("FindByToken", ctx, validToken).Return(mockVerifData, nil)

		expectedError := customerrors.ErrDataNotFound
		mockUserRepo.On("FindByID", ctx, userID).Return(nil, expectedError)

		serviceAuth := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil)

		err := serviceAuth.VerifyUser(ctx, validToken)

		assert.ErrorIs(t, err, expectedError)
		mockAuthRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)

		mockUserRepo.AssertNotCalled(t, "VerifUser")
	})

	// failed scenario: account already verified
	t.Run("Failed - Account Already Verified", func(t *testing.T) {
		ctx := context.Background()
		token := "TOKEN-123"
		userID := "USER-123"

		mockVerifData := &entity.UserVerification{
			UserID: userID,
			Token: token,
			ExpiredAt: time.Now().Add(1 * time.Hour),
		}

		mockAuthRepo := new(MockUserVerificationRepo)
		mockUserRepo := new(MockUserRepo)

		mockAuthRepo.On("FindByToken", ctx, token).Return(mockVerifData, nil)

		mockUserData := &entity.User{
			IsVerified: true,
		}
		mockUserData.ID = userID

		mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil)

		serviceAuth := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil)
		err := serviceAuth.VerifyUser(ctx, token)

		assert.ErrorIs(t, err, customerrors.ErrAccountAlreadyVerified)
		mockAuthRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockUserRepo.AssertNotCalled(t, "VerifUser")
	})
}
