package http_test

import (
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	httpDelivery "be-ayaka/internal/delivery/http"
	"be-ayaka/internal/middleware"
	mocksPkg "be-ayaka/internal/testingutils/mocks/pkg"
	mocksService "be-ayaka/internal/testingutils/mocks/service"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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

/////////////////// HELPER FUNCTION TO MAKE REQUEST ///////////////////
func (s *AuthHandlerSuite) makeRequest(method, path string, body interface{}) *http.Request {
    var bodyReader io.Reader
    if body != nil {
        if b, ok := body.([]byte); ok {
            bodyReader = bytes.NewBuffer(b)
        } else {
            marshaled, _ := json.Marshal(body)
            bodyReader = bytes.NewBuffer(marshaled)
        }
    }

    req, _ := http.NewRequest(method, path, bodyReader)
    req.Host = "localhost"
    req.Header.Set("Content-Type", "application/json")
    
    return req
}



/////////////////// TEST REGISTER USER ///////////////////
// 1. success scenario
func (s *AuthHandlerSuite) TestRegisterUser_Success() {
	requestBody := entity.UserRequest{
		Username:    "riku",
		Email:       "riku@mail.com",
		DisplayName: "Riku",
		Password:    "P4$$w0rd",
	}

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockService.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

	resp, err := s.app.Test(s.makeRequest("POST", "/api/v1/auth/register", requestBody))

	s.Require().NoError(err)
	s.Equal(fiber.StatusCreated, resp.StatusCode)

	s.mockValidator.AssertExpectations(s.T())
	s.mockService.AssertExpectations(s.T())
}

// 2. failed scenario: bad request
func (s *AuthHandlerSuite) TestRegisterUser_Failed_BadRequest() {
	body := []byte(`{"username": "riku", "email": "riku@mail.com"`)

	resp, err := s.app.Test(s.makeRequest("POST", "/api/v1/auth/register", body))

	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)

	s.mockValidator.AssertNotCalled(s.T(), "Validate", mock.Anything, mock.Anything)
	s.mockService.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
}

// 3. failed scenario: error validation
func (s *AuthHandlerSuite) TestRegisterUser_Failed_ValidationError() {
	requestBody := entity.UserRequest{Username: "ri"}

	expectedErr := customerrors.NewValidationError(
		`"username": "username must be at least 3 characters"`,
	)

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).
		Return(expectedErr).Once()

	resp, err := s.app.Test(s.makeRequest("POST", "/api/v1/auth/register", requestBody))

	s.Require().NoError(err)
	s.Equal(fiber.StatusUnprocessableEntity, resp.StatusCode)

	s.mockService.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
	s.mockValidator.AssertExpectations(s.T())
}

// 4. failed scenario: internal server error
func (s *AuthHandlerSuite) TestRegisterUser_Failed_InternalServerError() {
	requestBody := entity.UserRequest{
		Username: "riku",
		Email:    "riku@mail.com",
	}

	s.mockValidator.On("Validate", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockService.On("Create", mock.Anything, mock.Anything).
		Return(errors.New("failed to create user")).Once()

	resp, err := s.app.Test(s.makeRequest("POST", "/api/v1/auth/register", requestBody))

	s.Require().NoError(err)
	s.Equal(fiber.StatusInternalServerError, resp.StatusCode)

	s.mockValidator.AssertExpectations(s.T())
	s.mockService.AssertExpectations(s.T())
}
/////////////////// TEST REGISTER USER ///////////////////