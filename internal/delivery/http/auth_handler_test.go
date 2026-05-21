package http_test

import (
	"be-ayaka/internal/core/customerrors"
	httpDelivery "be-ayaka/internal/delivery/http"
	"be-ayaka/internal/delivery/http/dto"
	"be-ayaka/internal/middleware"
	"be-ayaka/internal/testingutils"
	mocksPkg "be-ayaka/internal/testingutils/mocks/pkg"
	mocksService "be-ayaka/internal/testingutils/mocks/service"
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AuthHandlerSuite struct {
	suite.Suite
	app           *fiber.App
	mockService   *mocksService.MockAuthService
	mockValidator *mocksPkg.MockValidator
	handler       *httpDelivery.AuthHandler
}

func (s *AuthHandlerSuite) SetupTest() {
	s.app = fiber.New(fiber.Config{
		ErrorHandler: middleware.GlobalErrorHandler,
	})
	s.mockService = new(mocksService.MockAuthService)
	s.mockValidator = new(mocksPkg.MockValidator)
	s.handler = httpDelivery.NewAuthHandler(s.mockService, s.mockValidator)

	s.app.Post("/api/v1/auth/register", s.handler.RegisterUser)
	s.app.Post("/api/v1/auth/resend-verification", s.handler.ResendVerification)
	s.app.Get("/api/v1/auth/verify", s.handler.VerifyEmail)
	s.app.Post("/api/v1/auth/login", s.handler.Login)
	s.app.Post("/api/v1/auth/refresh", s.handler.RefreshToken)
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerSuite))
}

// ///////////////// TEST REGISTER USER ///////////////////
// 1. success scenario
func (s *AuthHandlerSuite) TestRegisterUser_Success() {
	requestBody := dto.UserRequest{
		Username:    "riku",
		Email:       "riku@mail.com",
		DisplayName: "Riku",
		Password:    "P4$$w0rd",
	}

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockService.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

	resp, err := s.app.Test(testingutils.MakeJSONRequest("POST", "/api/v1/auth/register", requestBody))

	s.Require().NoError(err)
	s.Equal(fiber.StatusCreated, resp.StatusCode)

	s.mockValidator.AssertExpectations(s.T())
	s.mockService.AssertExpectations(s.T())
}

// 2. failed scenario: bad request
func (s *AuthHandlerSuite) TestRegisterUser_Failed_BadRequest() {
	body := []byte(`{"username": "riku", "email": "riku@mail.com"`)

	resp, err := s.app.Test(testingutils.MakeJSONRequest("POST", "/api/v1/auth/register", body))

	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)

	s.mockValidator.AssertNotCalled(s.T(), "Validate", mock.Anything, mock.Anything)
	s.mockService.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
}

// 3. failed scenario: error validation
func (s *AuthHandlerSuite) TestRegisterUser_Failed_ValidationError() {
	requestBody := dto.UserRequest{Username: "ri"}

	expectedErr := customerrors.NewValidationError(
		`"username": "username must be at least 3 characters"`,
	)

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).
		Return(expectedErr).Once()

	resp, err := s.app.Test(testingutils.MakeJSONRequest("POST", "/api/v1/auth/register", requestBody))

	s.Require().NoError(err)
	s.Equal(fiber.StatusUnprocessableEntity, resp.StatusCode)

	s.mockService.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
	s.mockValidator.AssertExpectations(s.T())
}

// 4. failed scenario: internal server error
func (s *AuthHandlerSuite) TestRegisterUser_Failed_InternalServerError() {
	requestBody := dto.UserRequest{
		Username: "riku",
		Email:    "riku@mail.com",
	}

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockService.On("Create", mock.Anything, mock.Anything).
		Return(errors.New("failed to create user")).Once()

	resp, err := s.app.Test(testingutils.MakeJSONRequest("POST", "/api/v1/auth/register", requestBody))

	s.Require().NoError(err)
	s.Equal(fiber.StatusInternalServerError, resp.StatusCode)

	s.mockValidator.AssertExpectations(s.T())
	s.mockService.AssertExpectations(s.T())
}

/////////////////// TEST REGISTER USER ///////////////////

// ///////////////// TEST LOGIN USER ///////////////////
// 1. success scenario
func (s *AuthHandlerSuite) TestLogin_Success() {
	requestBody := dto.LoginRequest{
		EmailUsername: "riku",
		Password:      "P4$$w0rd",
	}

	mockRes := &dto.LoginResponse{
		User: dto.UserResponse{
			Username: "riku",
		},
	}
	mockRes.AccessToken = "aaccess-token"
	mockRes.RefreshToken = "refresh-token"

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockService.On("Login", mock.Anything, requestBody.EmailUsername, requestBody.Password, mock.Anything).Return(mockRes, nil).Once()

	req := testingutils.MakeJSONRequest("POST", "/api/v1/auth/login", requestBody)
	res, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, res.StatusCode)

	s.mockValidator.AssertExpectations(s.T())
	s.mockService.AssertExpectations(s.T())
}

// 2. failed scenario: email or username not found
func (s *AuthHandlerSuite) TestLogin_Failed_BadRequest() {
	requestBody := []byte(`{"emailUsername": "riku", "password": "pass"`)

	req := testingutils.MakeJSONRequest("POST", "/api/v1/auth/login", requestBody)
	res, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, res.StatusCode)

	s.mockValidator.AssertNotCalled(s.T(), "Validate")
	s.mockService.AssertNotCalled(s.T(), "Login")
}

// 3. failed scenario: failed validation
func (s *AuthHandlerSuite) TestLogin_Failed_ValidationError() {
	requestBody := dto.LoginRequest{
		EmailUsername: "",
		Password:      "P4$$w0rd",
	}

	expectedErr := customerrors.NewValidationError(
		`"emailUsername": "emailUsername is required"`,
	)

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).Return(expectedErr).Once()

	req := testingutils.MakeJSONRequest("POST", "/api/v1/auth/login", requestBody)
	res, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusUnprocessableEntity, res.StatusCode)

	s.mockService.AssertNotCalled(s.T(), "Login")
	s.mockValidator.AssertExpectations(s.T())
}

// 4. FAILED scenario: internal server error
func (s *AuthHandlerSuite) TestLogin_Failed_InternalServerError() {
	requestBody := dto.LoginRequest{
		EmailUsername: "riku",
		Password:      "P4$$w0rd",
	}

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockService.On("Login", mock.Anything, requestBody.EmailUsername, requestBody.Password, mock.Anything).
		Return((*dto.LoginResponse)(nil), errors.New("failed to login")).Once()

	req := testingutils.MakeJSONRequest("POST", "/api/v1/auth/login", requestBody)
	res, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusInternalServerError, res.StatusCode)

	s.mockValidator.AssertExpectations(s.T())
	s.mockService.AssertExpectations(s.T())
}

/////////////////// TEST LOGIN USER ///////////////////

// ///////////////// TEST RESEND VERIFICATION USER ///////////////////
// 1. success scenario
func (s *AuthHandlerSuite) TestResendVerification_Success() {
	requestBody := dto.UserVerificationRequest{
		Email: "riku@mail.com",
	}

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockService.On("ResendVerification", mock.Anything, requestBody.Email).Return(nil).Once()

	req := testingutils.MakeJSONRequest("POST", "/api/v1/auth/resend-verification", requestBody)
	res, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, res.StatusCode)

	s.mockValidator.AssertExpectations(s.T())
	s.mockService.AssertExpectations(s.T())
}

// 2. failed scenario: bad request
func (s *AuthHandlerSuite) TestResendVerification_Failed_BadRequest() {
	requestBody := []byte(`{"email": "riku@mail.com"`)

	req := testingutils.MakeJSONRequest("POST", "/api/v1/auth/resend-verification", requestBody)
	res, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, res.StatusCode)

	s.mockValidator.AssertNotCalled(s.T(), "Validate")
	s.mockService.AssertNotCalled(s.T(), "ResendVerification")
}

// 3. failed scenario: failed validation
func (s *AuthHandlerSuite) TestResendVerification_Failed_ValidationError() {
	requestBody := dto.UserVerificationRequest{
		Email: "",
	}

	expectedErr := customerrors.NewValidationError(
		`"email": "email is required"`,
	)

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).Return(expectedErr).Once()

	req := testingutils.MakeJSONRequest("POST", "/api/v1/auth/resend-verification", requestBody)
	res, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusUnprocessableEntity, res.StatusCode)

	s.mockValidator.AssertExpectations(s.T())
	s.mockService.AssertNotCalled(s.T(), "ResendVerification")
}

// 4. failed scenario: internal server error
func (s *AuthHandlerSuite) TestResendVerification_Failed_InternalServerError() {
	requestBody := dto.UserVerificationRequest{
		Email: "riku@mail.com",
	}

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockService.On("ResendVerification", mock.Anything, requestBody.Email).Return(errors.New("failed to resend verification")).Once()

	req := testingutils.MakeJSONRequest("POST", "/api/v1/auth/resend-verification", requestBody)
	res, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusInternalServerError, res.StatusCode)

	s.mockValidator.AssertExpectations(s.T())
	s.mockService.AssertExpectations(s.T())
}

/////////////////// TEST RESEND VERIFICATION USER ///////////////////

// ///////////////// TEST VERIFY EMAIL USER ///////////////////
// 1. success scenario
func (s *AuthHandlerSuite) TestVerifyEmail_Success() {
	validToken := "TOKEN-123"

	path := "/api/v1/auth/verify?token=" + validToken

	s.mockService.On("VerifyUser", mock.Anything, validToken).Return(nil).Once()

	req := testingutils.MakeJSONRequest("GET", path, nil)
	resp, err := s.app.Test(req)

	// Assertions
	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, resp.StatusCode)

	s.mockService.AssertExpectations(s.T())
}

// 2. failed scenario: token missing
func (s *AuthHandlerSuite) TestVerifyEmail_FailedTokenMissing() {

	req := testingutils.MakeJSONRequest("GET", "/api/v1/auth/verify", nil)
	resp, err := s.app.Test(req)

	// Assertions
	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)

	s.mockService.AssertNotCalled(s.T(), "VerifyUser", mock.Anything, mock.Anything)
}

// 3. failed scenario: failed to verify email
func (s *AuthHandlerSuite) TestVerifyEmail_FailedVerifyToken() {
	token := "TOKEN-123"
	expectedErr := errors.New("error verify")
	path := "/api/v1/auth/verify?token=" + token

	s.mockService.On("VerifyUser", mock.Anything, token).Return(expectedErr).Once()

	req := testingutils.MakeJSONRequest("GET", path, nil)
	resp, err := s.app.Test(req)

	// Assertions
	s.Require().NoError(err)
	s.Equal(fiber.StatusInternalServerError, resp.StatusCode)

	s.mockService.AssertExpectations(s.T())
}

/////////////////// TEST VERIFY EMAIL USER ///////////////////

/////////////////// TEST RESEND VERIFICATION USER ///////////////////
