package service_test

import (
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/service"
	"be-ayaka/internal/testingutils"
	"be-ayaka/internal/testingutils/mocks"
	"be-ayaka/pkg/jwt"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

		serviceAuth := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil, nil)
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

		serviceAuth := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil, nil)
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

		serviceAuth := service.NewAuthService(MockUserRepo, nil, mockAuthRepo, nil, nil, nil, nil)

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

		serviceAuth := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil, nil)

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

		serviceAuth := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil, nil)
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

		service := service.NewAuthService(mockUserRepo, mockHash, mockAuthRepo, mockEmail, testingutils.GetDummyConfig(), mockTx, nil)

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

		service := service.NewAuthService(mockUserRepo, mockHash, nil, nil, nil, mockTx, nil)
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

		service := service.NewAuthService(mockUserRepo, mockHash, mockAuthRepo, nil, nil, mockTx, nil)
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

		service := service.NewAuthService(mockUserRepo, mockHash, mockAuthRepo, mockEmail, testingutils.GetDummyConfig(), mockTx, nil)

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
			Email:      email,
			Username:   "yuriku",
		}
		mockUserData.ID = userID

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)
		mockEmail := new(mocks.MockEmail)

		mockAuthRepo.On("FindByEmail", ctx, email).Return(mockVerifData, nil)
		mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil)

		mockAuthRepo.On("Upsert", ctx, mock.Anything).Return(nil)
		mockEmail.On("SendEmail", email, mock.Anything, mock.Anything).Return(nil).Maybe()

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, mockEmail, testingutils.GetDummyConfig(), nil, nil)
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

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil, nil)
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

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil, nil)
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

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, nil, nil, nil, nil)
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
			UserID:    userID,
		}

		mockUserData := &entity.User{
			IsVerified: true,
		}

		mockAuthRepo := new(mocks.MockUserVerificationRepo)
		mockUserRepo := new(mocks.MockUserRepo)
		mockEmail := new(mocks.MockEmail)

		mockAuthRepo.On("FindByEmail", ctx, email).Return(mockAuthData, nil)
		mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil)

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, mockEmail, nil, nil, nil)
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

		service := service.NewAuthService(mockUserRepo, nil, mockAuthRepo, mockEmail, testingutils.GetDummyConfig(), nil, nil)
		err := service.ResendVerification(ctx, email)

		assert.ErrorIs(t, err, dbError)
		mockEmail.AssertNotCalled(t, "SendEmail")
		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})
}

// =============================== test login
func TestLogin(t *testing.T) {
	// 1. success scneario
	t.Run("Success - Login User", func(t *testing.T) {
		ctx := context.Background()
		emailUsn := "yuriku@mail.com"
		password := "password123"
		reqID := "REQ-123"
		hashedPassword := "hashedpassword"

		mockUser := &entity.User{
			Username:   "yuriku",
			Email:      emailUsn,
			Password:   hashedPassword,
			Role:       "User",
			IsVerified: true,
		}
		mockUser.ID = "USER-123"

		mockToken := &jwt.TokenPair{
			AccessToken:  "accessToken",
			RefreshToken: "refreshToken",
		}

		mockUserRepo := new(mocks.MockUserRepo)
		mockHash := new(mocks.MockHashService)
		mockTokenRepo := new(mocks.MockTokenService)

		mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return(mockUser, nil)
		mockHash.On("ComparePassword", hashedPassword, password).Return(nil)
		mockUserRepo.On("UpdateRefreshToken", ctx, mockUser.ID, mock.Anything).Return(nil)
		mockTokenRepo.On("GenerateToken", testingutils.GetDummyConfig(), mockUser.ID, mockUser.Role).Return(mockToken, nil)

		service := service.NewAuthService(mockUserRepo, mockHash, nil, nil, testingutils.GetDummyConfig(), nil, mockTokenRepo)
		res, err := service.Login(ctx, emailUsn, password, reqID)

		assert.NoError(t, err)
		assert.NotNil(t, res)

		assert.Equal(t, mockUser.ID, res.User.ID)
		assert.Equal(t, mockUser.Username, res.User.Username)
		assert.Equal(t, mockUser.Email, res.User.Email)
		assert.Equal(t, mockUser.Role, res.User.Role)

		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)

		mockUserRepo.AssertExpectations(t)
		mockHash.AssertExpectations(t)
	})

	// 2. failed scenario: user not found
	t.Run("Failed - User Not Found", func(t *testing.T) {
		ctx := context.Background()
		emailUsn := "riku@mail.com"
		password := "password"
		reqId := "REQ-123"

		mockUserRepo := new(mocks.MockUserRepo)
		mockHash := new(mocks.MockHashService)

		mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return(nil, customerrors.ErrDataNotFound)

		service := service.NewAuthService(mockUserRepo, mockHash, nil, nil, nil, nil, nil)
		res, err := service.Login(ctx, emailUsn, password, reqId)

		assert.Error(t, err)
		assert.ErrorIs(t, err, customerrors.ErrDataNotFound)

		assert.Nil(t, res)
		mockHash.AssertNotCalled(t, "ComparePassword")
		mockUserRepo.AssertNotCalled(t, "UpdateRefreshToken")
		mockUserRepo.AssertExpectations(t)
		mockHash.AssertExpectations(t)
	})

	// 3. failed scenario: user not verified
	t.Run("Failed - User Not Verified", func(t *testing.T) {
		ctx := context.Background()
		emailUsn := "riku@mail.com"
		password := "password"
		reqId := "REQ-123"
		hashedPassword := "hashedpassword"

		mockUser := &entity.User{
			Username:   "yuriku",
			Email:      emailUsn,
			Password:   hashedPassword,
			Role:       "User",
			IsVerified: false,
		}
		mockUser.ID = "USER-123"

		mockUserRepo := new(mocks.MockUserRepo)
		mockHash := new(mocks.MockHashService)

		mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return(mockUser, nil)

		service := service.NewAuthService(mockUserRepo, mockHash, nil, nil, nil, nil, nil)
		res, err := service.Login(ctx, emailUsn, password, reqId)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, customerrors.ErrAccountInactive)

		mockHash.AssertNotCalled(t, "ComparePassword")
		mockUserRepo.AssertNotCalled(t, "UpdateRefreshToken")
		mockHash.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	// 4. failed scenario: invalid password
	t.Run("Failed - Invalid Password", func(t *testing.T) {
		ctx := context.Background()
		emailUsn := "riku@mail.com"
		password := "password"
		reqId := "REQ-123"
		hashedPassword := "hashedpassword"

		mockUser := &entity.User{
			Username:   "yuriku",
			Email:      emailUsn,
			Password:   hashedPassword,
			Role:       "User",
			IsVerified: true,
		}
		mockUser.ID = "USER-123"

		mockUserRepo := new(mocks.MockUserRepo)
		mockHash := new(mocks.MockHashService)

		mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return(mockUser, nil)
		mockHash.On("ComparePassword", hashedPassword, password).Return(customerrors.ErrInvalidPassword)

		service := service.NewAuthService(mockUserRepo, mockHash, nil, nil, nil, nil, nil)
		res, err := service.Login(ctx, emailUsn, password, reqId)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, customerrors.ErrInvalidPassword)

		mockUserRepo.AssertNotCalled(t, "UpdateRefreshToken")
		mockHash.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	// 5. failed scenario: failed to generate token
	t.Run("Failed - Failed Generate Token", func(t *testing.T) {
		ctx := context.Background()
		emailUsn := "riku@mail.com"
		password := "password"
		reqId := "REQ-123"
		hashedPassword := "hashedpassword"

		mockUser := &entity.User{
			Username:   "yuriku",
			Email:      emailUsn,
			Password:   hashedPassword,
			Role:       "User",
			IsVerified: true,
		}
		mockUser.ID = "USER-123"

		mockUserRepo := new(mocks.MockUserRepo)
		mockHash := new(mocks.MockHashService)
		mockToken := new(mocks.MockTokenService)
		expectedError := errors.New("error token")

		mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return(mockUser, nil)
		mockHash.On("ComparePassword", hashedPassword, password).Return(nil)
		mockToken.On("GenerateToken", testingutils.GetDummyConfig(), mockUser.ID, mockUser.Role).Return(nil, expectedError)

		service := service.NewAuthService(mockUserRepo, mockHash, nil, nil, testingutils.GetDummyConfig(), nil, mockToken)
		res, err := service.Login(ctx, emailUsn, password, reqId)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, expectedError)

		mockUserRepo.AssertNotCalled(t, "UpdateRefreshToken")
		mockHash.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	// 6. failed scenario: failed to update refresh token
	t.Run("Failed - Failed Update Refresh Token", func(t *testing.T) {
		ctx := context.Background()
		emailUsn := "riku@mail.com"
		password := "password"
		reqId := "REQ-123"
		hashedPassword := "hashedpassword"

		mockUser := &entity.User{
			Username:   "yuriku",
			Email:      emailUsn,
			Password:   hashedPassword,
			Role:       "User",
			IsVerified: true,
		}
		mockUser.ID = "USER-123"

		mockToken := &jwt.TokenPair{
			AccessToken:  "accessToken",
			RefreshToken: "refreshToken",
		}

		mockUserRepo := new(mocks.MockUserRepo)
		mockHash := new(mocks.MockHashService)
		mockTokenRepo := new(mocks.MockTokenService)
		expectedError := errors.New("error db")

		mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return(mockUser, nil)
		mockHash.On("ComparePassword", hashedPassword, password).Return(nil)
		mockTokenRepo.On("GenerateToken", testingutils.GetDummyConfig(), mockUser.ID, mockUser.Role).Return(mockToken, nil)
		mockUserRepo.On("UpdateRefreshToken", ctx, mockUser.ID, mockToken.RefreshToken).Return(expectedError)

		service := service.NewAuthService(mockUserRepo, mockHash, nil, nil, testingutils.GetDummyConfig(), nil, mockTokenRepo)
		res, err := service.Login(ctx, emailUsn, password, reqId)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, expectedError)

		mockHash.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}
