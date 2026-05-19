package service_test

import (
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/service"
	"be-ayaka/internal/testingutils"
	mocksPkg "be-ayaka/internal/testingutils/mocks/pkg"
	mocksRepo "be-ayaka/internal/testingutils/mocks/repository"
	"be-ayaka/pkg/jwt"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AuthServiceSuite struct {
	suite.Suite
	mockUserRepo  *mocksRepo.MockUserRepo
	mockVerifRepo *mocksRepo.MockUserVerificationRepo
	mockHash      *mocksPkg.MockHashService
	mockTx        *mocksRepo.MockTxManager
	mockEmail     *mocksPkg.MockEmail
	mockToken     *mocksPkg.MockTokenService
	service       service.AuthService
}

func (s *AuthServiceSuite) SetupTest() {
	s.mockUserRepo = new(mocksRepo.MockUserRepo)
	s.mockVerifRepo = new(mocksRepo.MockUserVerificationRepo)
	s.mockHash = new(mocksPkg.MockHashService)
	s.mockTx = new(mocksRepo.MockTxManager)
	s.mockEmail = new(mocksPkg.MockEmail)
	s.mockToken = new(mocksPkg.MockTokenService)

	s.service = service.NewAuthService(
		s.mockUserRepo,
		s.mockHash,
		s.mockVerifRepo,
		s.mockEmail,
		testingutils.GetDummyConfig(),
		s.mockTx,
		s.mockToken,
	)
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceSuite))
}

// =============================================================================
// VERIFY USER SCENARIOS
// =============================================================================

func (s *AuthServiceSuite) TestVerifyUser_Success() {
	ctx := context.Background()
	validToken := "TOKEN-123"
	userID := "USER-123"

	mockVerifData := &entity.UserVerification{
		UserID:    userID,
		Token:     validToken,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}
	mockUserData := &entity.User{IsVerified: false}
	mockUserData.ID = userID

	s.mockVerifRepo.On("FindByToken", ctx, validToken).Return(mockVerifData, nil).Once()
	s.mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil).Once()
	s.mockUserRepo.On("VerifUser", ctx, userID).Return(nil).Once()

	err := s.service.VerifyUser(ctx, validToken)

	s.NoError(err)
	s.mockVerifRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestVerifyUser_Failed_TokenExpired() {
	ctx := context.Background()
	expiredToken := "TOKEN-EXPIRED"

	mockVerifData := &entity.UserVerification{
		UserID:    "USER-123",
		Token:     expiredToken,
		ExpiredAt: time.Now().Add(-1 * time.Hour),
	}

	s.mockVerifRepo.On("FindByToken", ctx, expiredToken).Return(mockVerifData, nil).Once()

	err := s.service.VerifyUser(ctx, expiredToken)

	s.ErrorIs(err, customerrors.ErrTokenExpired)
	s.mockVerifRepo.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestVerifyUser_Failed_TokenNotFound() {
	ctx := context.Background()
	notFoundToken := "TOKEN-UNAVAIL"
	expectedError := customerrors.ErrDataNotFound

	s.mockVerifRepo.On("FindByToken", ctx, notFoundToken).Return((*entity.UserVerification)(nil), expectedError).Once()

	err := s.service.VerifyUser(ctx, notFoundToken)

	s.ErrorIs(err, expectedError)
	s.mockVerifRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertNotCalled(s.T(), "FindByID")
}

func (s *AuthServiceSuite) TestVerifyUser_Failed_UserNotFound() {
	ctx := context.Background()
	validToken := "TOKEN-123"
	userID := "USER-123"

	mockVerifData := &entity.UserVerification{
		UserID:    userID,
		Token:     validToken,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}
	expectedError := customerrors.ErrDataNotFound

	s.mockVerifRepo.On("FindByToken", ctx, validToken).Return(mockVerifData, nil).Once()
	s.mockUserRepo.On("FindByID", ctx, userID).Return((*entity.User)(nil), expectedError).Once()

	err := s.service.VerifyUser(ctx, validToken)

	s.ErrorIs(err, expectedError)
	s.mockVerifRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertNotCalled(s.T(), "VerifUser")
}

func (s *AuthServiceSuite) TestVerifyUser_Failed_AccountAlreadyVerified() {
	ctx := context.Background()
	token := "TOKEN-123"
	userID := "USER-123"

	mockVerifData := &entity.UserVerification{
		UserID:    userID,
		Token:     token,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}
	mockUserData := &entity.User{IsVerified: true}
	mockUserData.ID = userID

	s.mockVerifRepo.On("FindByToken", ctx, token).Return(mockVerifData, nil).Once()
	s.mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil).Once()

	err := s.service.VerifyUser(ctx, token)

	s.ErrorIs(err, customerrors.ErrAccountAlreadyVerified)
	s.mockVerifRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertNotCalled(s.T(), "VerifUser")
}

// =============================================================================
// CREATE USER SCENARIOS
// =============================================================================

func (s *AuthServiceSuite) TestCreate_Success() {
	ctx := context.Background()
	user := &entity.UserRequest{
		Username: "yuriku",
		Email:    "yuriku@mail.com",
		Password: "password123",
	}
	hashedPassword := "HASHEDPASSWORD"

	s.mockHash.On("HashPassword", user.Password).Return(hashedPassword, nil).Once()
	s.mockTx.On("WithTx", ctx, mock.Anything).Return(nil).Once()
	s.mockUserRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
	s.mockVerifRepo.On("Upsert", ctx, mock.Anything).Return(nil).Once()
	s.mockEmail.On("SendEmail", user.Email, mock.Anything, mock.Anything).Return(nil).Maybe()

	err := s.service.Create(ctx, user)

	s.NoError(err)
	s.mockHash.AssertExpectations(s.T())
	s.mockTx.AssertExpectations(s.T())
	s.mockVerifRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockEmail.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestCreate_Failed_HashPassword() {
	ctx := context.Background()
	user := &entity.UserRequest{
		Username: "yuriku",
		Email:    "yuriku@mail.com",
		Password: "password123",
	}
	expectedError := customerrors.ErrFailHash
	s.mockHash.On("HashPassword", user.Password).Return("", expectedError).Once()

	err := s.service.Create(ctx, user)

	s.ErrorIs(err, expectedError)
	s.mockTx.AssertNotCalled(s.T(), "WithTx", mock.Anything, mock.Anything)
	s.mockUserRepo.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
	s.mockHash.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestCreate_Failed_SaveToUserTable() {
	ctx := context.Background()
	user := &entity.UserRequest{
		Username: "yuriku",
		Email:    "yuriku@mail.com",
		Password: "password123",
	}
	hashedPassword := "HASHEDPASSWORD"
	dbError := errors.New("sql error")

	s.mockHash.On("HashPassword", user.Password).Return(hashedPassword, nil).Once()
	s.mockUserRepo.On("Create", ctx, mock.Anything).Return(dbError).Once()
	s.mockTx.On("WithTx", ctx, mock.Anything).Return(dbError).Once()

	err := s.service.Create(ctx, user)

	s.ErrorIs(err, dbError)
	s.mockVerifRepo.AssertNotCalled(s.T(), "Upsert", mock.Anything, mock.Anything)
	s.mockHash.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockTx.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestCreate_Failed_InsideGenerateAndSendVerif() {
	ctx := context.Background()
	user := &entity.UserRequest{
		Username: "yuriku",
		Email:    "yuriku@mail.com",
		Password: "password123",
	}
	hashedPassword := "HASHEDPASSWORD"
	dbError := errors.New("sql error")

	s.mockHash.On("HashPassword", user.Password).Return(hashedPassword, nil).Once()
	s.mockUserRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
	s.mockVerifRepo.On("Upsert", ctx, mock.Anything).Return(dbError).Once()
	s.mockTx.On("WithTx", ctx, mock.Anything).Return(dbError).Once()

	err := s.service.Create(ctx, user)

	s.ErrorIs(err, dbError)
	s.mockEmail.AssertNotCalled(s.T(), "SendEmail", mock.Anything, mock.Anything, mock.Anything)
	s.mockHash.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockVerifRepo.AssertExpectations(s.T())
	s.mockTx.AssertExpectations(s.T())
}

// =============================================================================
// RESEND VERIFICATION SCENARIOS
// =============================================================================

func (s *AuthServiceSuite) TestResendVerification_Success() {
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

	s.mockVerifRepo.On("FindByEmail", ctx, email).Return(mockVerifData, nil).Once()
	s.mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil).Once()
	s.mockVerifRepo.On("Upsert", ctx, mock.Anything).Return(nil).Once()
	s.mockEmail.On("SendEmail", email, mock.Anything, mock.Anything).Return(nil).Maybe()

	err := s.service.ResendVerification(ctx, email)

	s.NoError(err)
	s.mockVerifRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockEmail.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestResendVerification_Failed_EmailNotFound() {
	ctx := context.Background()
	email := "yuriku@mail.com"

	s.mockVerifRepo.On("FindByEmail", ctx, email).Return((*entity.UserVerification)(nil), customerrors.ErrDataNotFound).Once()

	err := s.service.ResendVerification(ctx, email)

	s.ErrorIs(err, customerrors.ErrDataNotFound)
	s.mockUserRepo.AssertNotCalled(s.T(), "FindByID", mock.Anything, mock.Anything)
	s.mockVerifRepo.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestResendVerification_Failed_CooldownActive() {
	ctx := context.Background()
	email := "yuriku@mail.com"

	mockAuthData := &entity.UserVerification{
		CreatedAt: time.Now().Add(-2 * time.Minute),
	}

	s.mockVerifRepo.On("FindByEmail", ctx, email).Return(mockAuthData, nil).Once()

	err := s.service.ResendVerification(ctx, email)

	s.ErrorIs(err, customerrors.ErrCooldownActive)
	s.mockUserRepo.AssertNotCalled(s.T(), "FindByID", mock.Anything, mock.Anything)
	s.mockVerifRepo.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestResendVerification_Failed_UserIDNotFound() {
	ctx := context.Background()
	email := "yuriku@mail.com"

	mockAuthData := &entity.UserVerification{
		UserID:    "USER-123",
		CreatedAt: time.Now().Add(-10 * time.Minute),
	}

	s.mockVerifRepo.On("FindByEmail", ctx, email).Return(mockAuthData, nil).Once()
	s.mockUserRepo.On("FindByID", ctx, mockAuthData.UserID).Return((*entity.User)(nil), customerrors.ErrDataNotFound).Once()

	err := s.service.ResendVerification(ctx, email)

	s.ErrorIs(err, customerrors.ErrDataNotFound)
	s.mockEmail.AssertNotCalled(s.T(), "SendEmail", mock.Anything, mock.Anything, mock.Anything)
	s.mockVerifRepo.AssertNotCalled(s.T(), "Upsert", mock.Anything, mock.Anything)
	s.mockVerifRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestResendVerification_Failed_UserAlreadyVerified() {
	ctx := context.Background()
	email := "yuriku@mail.com"
	userID := "USER-123"

	mockAuthData := &entity.UserVerification{
		CreatedAt: time.Now().Add(-10 * time.Minute),
		UserID:    userID,
	}
	mockUserData := &entity.User{IsVerified: true}

	s.mockVerifRepo.On("FindByEmail", ctx, email).Return(mockAuthData, nil).Once()
	s.mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil).Once()

	err := s.service.ResendVerification(ctx, email)

	s.ErrorIs(err, customerrors.ErrAccountAlreadyVerified)
	s.mockEmail.AssertNotCalled(s.T(), "SendEmail", mock.Anything, mock.Anything)
	s.mockVerifRepo.AssertNotCalled(s.T(), "Upsert", mock.Anything, mock.Anything)
	s.mockVerifRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestResendVerification_Failed_InsideGenerateAndSendVerif() {
	ctx := context.Background()
	email := "yuriku@mail.com"
	userID := "USER-123"
	dbError := errors.New("sql error")

	mockVerifData := &entity.UserVerification{
		UserID:    userID,
		CreatedAt: time.Now().Add(-6 * time.Minute),
	}
	mockUserData := &entity.User{IsVerified: false}
	mockUserData.ID = userID

	s.mockVerifRepo.On("FindByEmail", ctx, email).Return(mockVerifData, nil).Once()
	s.mockUserRepo.On("FindByID", ctx, userID).Return(mockUserData, nil).Once()
	s.mockVerifRepo.On("Upsert", ctx, mock.Anything).Return(dbError).Once()

	err := s.service.ResendVerification(ctx, email)

	s.ErrorIs(err, dbError)
	s.mockEmail.AssertNotCalled(s.T(), "SendEmail")
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockVerifRepo.AssertExpectations(s.T())
}

// =============================================================================
// LOGIN SCENARIOS
// =============================================================================

func (s *AuthServiceSuite) TestLogin_Success() {
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

	s.mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return(mockUser, nil).Once()
	s.mockHash.On("ComparePassword", hashedPassword, password).Return(nil).Once()
	s.mockUserRepo.On("UpdateRefreshToken", ctx, mockUser.ID, mock.Anything).Return(nil).Once()
	s.mockToken.On("GenerateToken", testingutils.GetDummyConfig(), mockUser.ID, mockUser.Role).Return(mockToken, nil).Once()

	res, err := s.service.Login(ctx, emailUsn, password, reqID)

	s.NoError(err)
	s.NotNil(res)
	s.Equal(mockUser.ID, res.User.ID)
	s.Equal(mockUser.Username, res.User.Username)
	s.Equal(mockUser.Email, res.User.Email)
	s.Equal(mockUser.Role, res.User.Role)
	s.NotEmpty(res.AccessToken)
	s.NotEmpty(res.RefreshToken)

	s.mockUserRepo.AssertExpectations(s.T())
	s.mockHash.AssertExpectations(s.T())
	s.mockToken.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestLogin_Failed_UserNotFound() {
	ctx := context.Background()
	emailUsn := "riku@mail.com"
	password := "password"
	reqId := "REQ-123"

	s.mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return((*entity.User)(nil), customerrors.ErrDataNotFound).Once()

	res, err := s.service.Login(ctx, emailUsn, password, reqId)

	s.Error(err)
	s.ErrorIs(err, customerrors.ErrDataNotFound)
	s.Nil(res)

	s.mockHash.AssertNotCalled(s.T(), "ComparePassword")
	s.mockUserRepo.AssertNotCalled(s.T(), "UpdateRefreshToken")
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestLogin_Failed_UserNotVerified() {
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

	s.mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return(mockUser, nil).Once()

	res, err := s.service.Login(ctx, emailUsn, password, reqId)

	s.Error(err)
	s.Nil(res)
	s.ErrorIs(err, customerrors.ErrAccountInactive)

	s.mockHash.AssertNotCalled(s.T(), "ComparePassword")
	s.mockUserRepo.AssertNotCalled(s.T(), "UpdateRefreshToken")
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestLogin_Failed_InvalidPassword() {
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

	s.mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return(mockUser, nil).Once()
	s.mockHash.On("ComparePassword", hashedPassword, password).Return(customerrors.ErrInvalidPassword).Once()

	res, err := s.service.Login(ctx, emailUsn, password, reqId)

	s.Error(err)
	s.Nil(res)
	s.ErrorIs(err, customerrors.ErrInvalidPassword)

	s.mockUserRepo.AssertNotCalled(s.T(), "UpdateRefreshToken")
	s.mockHash.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestLogin_Failed_GenerateToken() {
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
	expectedError := errors.New("error token")

	s.mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return(mockUser, nil).Once()
	s.mockHash.On("ComparePassword", hashedPassword, password).Return(nil).Once()
	s.mockToken.On("GenerateToken", testingutils.GetDummyConfig(), mockUser.ID, mockUser.Role).Return((*jwt.TokenPair)(nil), expectedError).Once()

	res, err := s.service.Login(ctx, emailUsn, password, reqId)

	s.Error(err)
	s.Nil(res)
	s.ErrorIs(err, expectedError)

	s.mockUserRepo.AssertNotCalled(s.T(), "UpdateRefreshToken")
	s.mockHash.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockToken.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestLogin_Failed_UpdateRefreshToken() {
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
	expectedError := errors.New("error db")

	s.mockUserRepo.On("FindByEmailUsername", ctx, emailUsn).Return(mockUser, nil).Once()
	s.mockHash.On("ComparePassword", hashedPassword, password).Return(nil).Once()
	s.mockToken.On("GenerateToken", testingutils.GetDummyConfig(), mockUser.ID, mockUser.Role).Return(mockToken, nil).Once()
	s.mockUserRepo.On("UpdateRefreshToken", ctx, mockUser.ID, mockToken.RefreshToken).Return(expectedError).Once()

	res, err := s.service.Login(ctx, emailUsn, password, reqId)

	s.Error(err)
	s.Nil(res)
	s.ErrorIs(err, expectedError)

	s.mockHash.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockToken.AssertExpectations(s.T())
}

// =============================================================================
// NEW ACCESS TOKEN SCENARIOS
// =============================================================================

func (s *AuthServiceSuite) TestNewAccessToken_Success() {
	ctx := context.Background()
	refreshToken := "REFRESH-123"
	requestId := "REQ-123"

	mockUser := &entity.User{Role: "user"}
	mockUser.ID = "USER-123"
	mockToken := &jwt.TokenPair{
		AccessToken:  "NEW_ACC_TOKEN",
		RefreshToken: "MEW_REFRESH_TOKEN",
	}

	s.mockUserRepo.On("FindByRefreshToken", ctx, refreshToken).Return(mockUser, nil).Once()
	s.mockToken.On("GenerateToken", testingutils.GetDummyConfig(), mockUser.ID, mockUser.Role).Return(mockToken, nil).Once()
	s.mockUserRepo.On("UpdateRefreshToken", ctx, mockUser.ID, mockToken.RefreshToken).Return(nil).Once()

	res, err := s.service.NewAccessToken(ctx, refreshToken, requestId)

	s.NoError(err)
	s.NotNil(res)
	s.Equal(mockToken.AccessToken, res.AccessToken)
	s.Equal(mockToken.RefreshToken, res.RefreshToken)

	s.mockUserRepo.AssertExpectations(s.T())
	s.mockToken.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestNewAccessToken_Failed_RefreshTokenEmpty() {
	ctx := context.Background()
	refreshToken := ""
	requestId := "REQ-123"

	res, err := s.service.NewAccessToken(ctx, refreshToken, requestId)

	s.Nil(res)
	s.Error(err)
	s.ErrorIs(err, customerrors.ErrUnauthorized)

	s.mockUserRepo.AssertNotCalled(s.T(), "FindByRefreshToken")
	s.mockUserRepo.AssertNotCalled(s.T(), "UpdateRefreshToken")
	s.mockToken.AssertNotCalled(s.T(), "GenerateToken")
}

func (s *AuthServiceSuite) TestNewAccessToken_Failed_RefreshTokenNotFound() {
	ctx := context.Background()
	refreshToken := "REFRESH-123"
	requestId := "REQ-123"

	s.mockUserRepo.On("FindByRefreshToken", ctx, refreshToken).Return((*entity.User)(nil), customerrors.ErrInvalidCredentials).Once()

	res, err := s.service.NewAccessToken(ctx, refreshToken, requestId)

	s.Nil(res)
	s.Error(err)
	s.ErrorIs(err, customerrors.ErrInvalidCredentials)

	s.mockToken.AssertNotCalled(s.T(), "GenerateToken")
	s.mockUserRepo.AssertNotCalled(s.T(), "UpdateRefreshToken")
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestNewAccessToken_Failed_GenerateNewToken() {
	ctx := context.Background()
	refreshToken := "REFRESH-123"
	requestId := "REQ-123"

	mockUser := &entity.User{Role: "user"}
	mockUser.ID = "USER-123"
	expectedErr := errors.New("failed create token")

	s.mockUserRepo.On("FindByRefreshToken", ctx, refreshToken).Return(mockUser, nil).Once()
	s.mockToken.On("GenerateToken", testingutils.GetDummyConfig(), mockUser.ID, mockUser.Role).Return((*jwt.TokenPair)(nil), expectedErr).Once()

	res, err := s.service.NewAccessToken(ctx, refreshToken, requestId)

	s.Nil(res)
	s.Error(err)
	s.ErrorIs(err, expectedErr)

	s.mockUserRepo.AssertNotCalled(s.T(), "UpdateRefreshToken")
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockToken.AssertExpectations(s.T())
}

func (s *AuthServiceSuite) TestNewAccessToken_Failed_UpdateRefreshToken() {
	ctx := context.Background()
	refreshToken := "REFRESH-123"
	requestId := "REQ-123"

	mockUser := &entity.User{Role: "user"}
	mockUser.ID = "USER-123"
	mockToken := &jwt.TokenPair{
		AccessToken:  "NEW_ACC_TOKEN",
		RefreshToken: "MEW_REFRESH_TOKEN",
	}
	expectedErr := errors.New("error db")

	s.mockUserRepo.On("FindByRefreshToken", ctx, refreshToken).Return(mockUser, nil).Once()
	s.mockToken.On("GenerateToken", testingutils.GetDummyConfig(), mockUser.ID, mockUser.Role).Return(mockToken, nil).Once()
	s.mockUserRepo.On("UpdateRefreshToken", ctx, mockUser.ID, mockToken.RefreshToken).Return(expectedErr).Once()

	res, err := s.service.NewAccessToken(ctx, refreshToken, requestId)

	s.Nil(res)
	s.Error(err)
	s.ErrorIs(err, expectedErr)

	s.mockUserRepo.AssertExpectations(s.T())
	s.mockToken.AssertExpectations(s.T())
}