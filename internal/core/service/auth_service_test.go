package service_test

import (
	"be-ayaka/config"
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/service"
	"be-ayaka/internal/mocks"
	"context"
	"errors"
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

	// failed scenario: failed to hash password
	t.Run("Failed-Failed to Hash Password", func(t *testing.T) {
		ctx := context.Background()
		user := &entity.UserRequest{
			Username: "yuriku",
			Email:    "yuriku@mail.com",
			Password: "password123",
		}
		mockHash := new(mocks.MockHashService)
		mockUserRepo := new(mocks.MockUserRepo)
		mockTx := new(mocks.MockTxManager)

		expectedError := customerrors.ErrFailHash
		mockHash.On("HashPassword", user.Password).Return("", expectedError)

		service := service.NewAuthService(mockUserRepo, mockHash, nil, nil, nil, mockTx)
		err := service.Create(ctx, user)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedError)

		mockTx.AssertNotCalled(t, "WithTx", mock.Anything, mock.Anything)
		mockUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)

		mockHash.AssertExpectations(t)
	})

	// failed scenario: failed to create user / save to table user
	t.Run("Failed - Failed Save Data to Table User", func(t *testing.T) {
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

		dbError := errors.New("sql error")

		mockHash.On("HashPassword", user.Password).Return(hashedPassword, nil)
		mockUserRepo.On("Create", ctx, mock.Anything).Return(dbError)
		mockTx.On("WithTx", ctx, mock.Anything).Return(dbError)

		service := service.NewAuthService(mockUserRepo, mockHash, mockAuthRepo, nil, nil, mockTx)
		err := service.Create(ctx, user)

		assert.Error(t, err)
		assert.ErrorIs(t, err, dbError)
		mockAuthRepo.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything)

		mockHash.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	// failed scenario: failed to save token to table verify user
	t.Run("Failed - Error Inside func generateAndSendVerif", func(t *testing.T) {
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

		dbError := errors.New("sql error")

		mockHash.On("HashPassword", user.Password).Return(hashedPassword, nil)
		mockUserRepo.On("Create", ctx, mock.Anything).Return(nil)
		mockAuthRepo.On("Upsert", ctx, mock.Anything).Return(dbError)
		mockTx.On("WithTx", ctx, mock.Anything).Return(dbError)

		service := service.NewAuthService(mockUserRepo, mockHash, mockAuthRepo, mockEmail, dummyCfg, mockTx)

		err := service.Create(ctx, user)

		assert.Error(t, err)
		assert.ErrorIs(t, err, dbError)

		mockEmail.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)

		mockHash.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

// =============================== test resend verification
func TestResendVerification(t *testing.T) {
	// success scneario
	t.Run("Success Resend Verification", func(t *testing.T) {
		ctx := context.Background()
		email := "yuriku@mail.com"
		userID := "USER-123"

		mockVerifData := &entity.UserVerification{
			UserID:    userID,
			CreatedAt: time.Now().Add(-6 * time.Minute),
		}

		mockUserData := &entity.User{
			IsVerified: false,
			Email: email,
			Username: "yuriku",
		}
		mockUserData.ID = userID

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)
		mockEmail := new(mocks.MockEmail)

		mockAuthRepo.On("FindByEmail", ctx, email).Return(mockVerifData, nil)
		mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil)

		mockAuthRepo.On("Upsert", ctx, mock.Anything).Return(nil)
		mockEmail.On("SendEmail", email, mock.Anything, mock.Anything).Return(nil).Maybe()

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, mockEmail, dummyCfg, nil)
		err := service.ResendVerification(ctx, email)

		assert.NoError(t, err)
		mockAuthRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockEmail.AssertExpectations(t)
	})

	// failed scenario: email not found
	t.Run("Failed - Email Not Found", func(t *testing.T) {
		ctx := context.Background()
		email := "yuriku@mail.com"

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)

		mockAuthRepo.On("FindByEmail", ctx, email).Return(nil, customerrors.ErrDataNotFound)

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil)
		err := service.ResendVerification(ctx, email)

		assert.ErrorIs(t, err, customerrors.ErrDataNotFound)
		mockUserRepo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
		mockAuthRepo.AssertExpectations(t)
	})

	// failed scenario: cooldown still active
	t.Run("Failed - Cooldown is active", func(t *testing.T) {
		ctx := context.Background()
		email := "yuriku@mail.com"

		mockAuthData := &entity.UserVerification{
			CreatedAt: time.Now().Add(-2 * time.Minute),
		}

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)

		mockAuthRepo.On("FindByEmail", ctx, email).Return(mockAuthData, nil)

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil)
		err := service.ResendVerification(ctx, email)

		assert.ErrorIs(t, err, customerrors.ErrCooldownActive)
		mockUserRepo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
		mockAuthRepo.AssertExpectations(t)
	})

	// failed scenario: user id not found or error on find by id function
	t.Run("Failed - User ID Not Found", func(t *testing.T) {
		ctx := context.Background()
		email := "yuriku@mail.com"

		mockAuthData := &entity.UserVerification{
			UserID:    "USER-123",
			CreatedAt: time.Now().Add(-10 * time.Minute),
		}

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)
		mockEmail := new(mocks.MockEmail)

		mockAuthRepo.On("FindByEmail", ctx, email).Return(mockAuthData, nil)
		mockUserRepo.On("FindByID", ctx, mockAuthData.UserID).Return(nil, customerrors.ErrDataNotFound)

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil)
		err := service.ResendVerification(ctx, email)

		assert.ErrorIs(t, err, customerrors.ErrDataNotFound)
		mockEmail.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
		mockAuthRepo.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything)
		mockAuthRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	// failed scenario: user already verified
	t.Run("Failed - User Already Verified", func(t *testing.T) {
		ctx := context.Background()
		email := "yuriku@mail.com"
		userID := "USER-123"

		mockAuthData := &entity.UserVerification{
			CreatedAt: time.Now().Add(-10 * time.Minute),
			UserID: userID,
		}

		mockUserData := &entity.User{
			IsVerified: true,
		}

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)
		mockEmail := new(mocks.MockEmail)

		mockAuthRepo.On("FindByEmail", ctx, email).Return(mockAuthData, nil)
		mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil)

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, mockEmail, nil, nil)
		err := service.ResendVerification(ctx, email)

		assert.ErrorIs(t, err, customerrors.ErrAccountAlreadyVerified)
		mockEmail.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything)
		mockAuthRepo.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything)
		mockAuthRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	// failed scenario: failed to save token token to table verify user
	t.Run("Failed - Error inside func generateAndSendVerif", func(t *testing.T) {
		ctx := context.Background()
		email := "yuriku@mail.com"
		userID := "ISER-123"

		mockVerifData := &entity.UserVerification{
			UserID:    userID,
			CreatedAt: time.Now().Add(-6 * time.Minute),
		}

		mockUserData := &entity.User{
			IsVerified: false,
		}
		mockUserData.ID = userID

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)
		mockEmail := new(mocks.MockEmail)

		dbError := errors.New("sql error")

		mockAuthRepo.On("FindByEmail", ctx, email).Return(mockVerifData, nil)
		mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil)

		mockAuthRepo.On("Upsert", ctx, mock.Anything).Return(dbError)

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, mockEmail, dummyCfg, nil)
		err := service.ResendVerification(ctx, email)

		assert.ErrorIs(t, err, dbError)
		mockEmail.AssertNotCalled(t, "SendEmail")
		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})
}