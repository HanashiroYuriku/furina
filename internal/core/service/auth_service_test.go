package service_test

import (
	"be-ayaka/config"
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/service"
	"be-ayaka/internal/mocks"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// var dummy
var dummyCfg = &config.Config{
	Frontend: config.FrontendConfig{
		URL: "http://localhost:3000",
	},
}

// =============================== test verify user
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
		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)

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

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)

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

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		MockUserRepo := new(mocks.MockUserRepo)

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
			UserID:    userID,
			Token:     validToken,
			ExpiredAt: time.Now().Add(1 * time.Hour),
		}

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)

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
			UserID:    userID,
			Token:     token,
			ExpiredAt: time.Now().Add(1 * time.Hour),
		}

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)

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

// =============================== test create
func TestCreate(t *testing.T) {
	// scenario: success create user
	t.Run("Success Create", func(t *testing.T) {
		ctx := context.Background()
		user := &entity.UserRequest{
			Username: "yuriku",
			Email:    "yuriku@mail.com",
			Password: "password123",
		}
		hashedPassword := "HASHEDPASSWORD"

		mockHash := new(mocks.MockHashService)
		mockTx := new(mocks.MockTxManager)
		mockUserRepo := new(mocks.MockUserRepo)
		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockEmail := new(mocks.MockEmail)

		mockHash.On("HashPassword", user.Password).Return(hashedPassword, nil)
		mockTx.On("WithTx", ctx, mock.Anything).Return(nil)
		mockUserRepo.On("Create", ctx, mock.Anything).Return(nil)
		mockAuthRepo.On("Upsert", ctx, mock.Anything).Return(nil)
		mockEmail.On("SendEmail", user.Email, mock.Anything, mock.Anything).Return(nil).Maybe()

		service := service.NewAuthService(mockUserRepo, mockHash, mockAuthRepo, mockEmail, dummyCfg, mockTx)

		err := service.Create(ctx, user)

		assert.NoError(t, err)
		mockHash.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockEmail.AssertExpectations(t)
	})

	// failed scenario: 
}
