package http

import (
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/service"
	"be-ayaka/internal/delivery/http/dto"
	"be-ayaka/pkg/logger"
	"be-ayaka/pkg/requestid"
	"be-ayaka/pkg/response"
	"be-ayaka/pkg/validator"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService service.AuthService
	validator   validator.Validator
}

func NewAuthHandler(authService service.AuthService, validator validator.Validator) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
	}
}

// RegisterUser register new user
// @Summary Register new User
// @Description Register new user using email, username, and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.UserRequest true "Payload Register"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 422 {object} response.Response "Validation Failed"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) RegisterUser(c *fiber.Ctx) error {
	requestId := requestid.GetRequestID(c)

	var request dto.UserRequest

	if err := c.BodyParser(&request); err != nil {
		go logger.Log("SYSTEM", "ERROR", "Failed to parse request body: "+err.Error(), requestId)
		return customerrors.ErrBadRequest
	}

	if err := h.validator.Validate(c.Context(), &request); err != nil {
		go logger.Log("SYSTEM", "WARN", "Validation failed: "+err.Error(), requestId)
		return err
	}

	if err := h.authService.Create(c.Context(), &request); err != nil {
		go logger.Log("SYSTEM", "ERROR", "Internal Server Error: "+err.Error(), requestId)
		return err
	}

	go logger.Log("SYSTEM", "INFO", fmt.Sprintf("User %s created successfully", request.Username), requestId)
	return c.Status(fiber.StatusCreated).JSON(
		response.NewSuccessResponse(
			response.StatusSuccess,
			"Success Create User",
			nil,
			requestId,
		),
	)
}

// ResendVerification resend verification email to user
// @Summary Resend new verification email to user
// @Description Resend new email verification to user using email after user failed to receive or expired token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.UserVerificationRequest true "Payload Email Verification"
// @Success 200 {object} response.Response{data=nil}
// @Failure 400 {object} response.Response
// @Failure 422 {object} response.Response "Validation Failed"
// @Failure 404 {object} response.Response "Data Not Found"
// @Router /api/v1/auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *fiber.Ctx) error {
	requestId := requestid.GetRequestID(c)

	var email dto.UserVerificationRequest
	if err := c.BodyParser(&email); err != nil {
		go logger.Log("SYSTEM", "ERROR", "Failed to parse request body: "+err.Error(), requestId)
		return customerrors.ErrBadRequest
	}

	if err := h.validator.Validate(c.Context(), &email); err != nil {
		go logger.Log("SYSTEM", "WARN", "Validation failed: "+err.Error(), requestId)
		return err
	}

	if err := h.authService.ResendVerification(c.Context(), email.Email); err != nil {
		go logger.Log("SYSTEM", "ERROR", "Internal Server Error: "+err.Error(), requestId)
		return err
	}

	go logger.Log("SYSTEM", "INFO", fmt.Sprintf("Verification email resent to %s successfully", email.Email), requestId)
	return c.Status(fiber.StatusOK).JSON(
		response.NewSuccessResponse(
			response.StatusSuccess,
			"Success Resend Verification Email",
			nil,
			requestId,
		),
	)
}

// VerifyEmail verify user's email
// @Summary verify user's email
// @Description verifies user's email using token from email verif link and activate the account
// @Tags Authentication
// @Produce json
// @Param token query string true "Verify Email Token"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "Unauthorized/Token Expired"
// @Failure 409 {object} response.Response "Data Conflict"
// @Failure 404 {object} response.Response "Data Not Found"
// @Router /api/v1/auth/verify [get]
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	requestId := requestid.GetRequestID(c)

	token := c.Query("token")
	if token == "" {
		go logger.Log("SYSTEM", "WARN", "Token query parameter is missing", requestId)
		return customerrors.ErrBadRequest
	}

	if err := h.authService.VerifyUser(c.Context(), token); err != nil {
		go logger.Log("SYSTEM", "ERROR", "Verification failed for token "+token+": "+err.Error(), requestId)
		return err
	}

	go logger.Log("SYSTEM", "INFO", "User email verified successfully", requestId)
	return c.Status(fiber.StatusOK).JSON(
		response.NewSuccessResponse(
			response.StatusSuccess,
			"Email Verified Successfully",
			nil,
			requestId,
		),
	)
}

// Login authenticate user and generate jwt tokens
// @Summary Login User
// @Description Login using email or username and password and receive access & refresh tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Payload Login"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 422 {object} response.Response "Validation Failed"
// @Failure 404 {object} response.Response "Email/username does not exist"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	requestId := requestid.GetRequestID(c)

	var request dto.LoginRequest
	if err := c.BodyParser(&request); err != nil {
		go logger.Log("SYSTEM", "ERROR", "Failed to parse request body: "+err.Error(), requestId)
		return customerrors.ErrBadRequest
	}

	if err := h.validator.Validate(c.Context(), &request); err != nil {
		go logger.Log("SYSTEM", "WARN", "Validation failed: "+err.Error(), requestId)
		return err
	}

	res, err := h.authService.Login(c.Context(), request.EmailUsername, request.Password, requestId)
	if err != nil {
		go logger.Log("SYSTEM", "ERROR", "Login failed for user "+request.EmailUsername+": "+err.Error(), requestId)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(
		response.StatusSuccess,
		"Login Success",
		res,
		requestId,
	))
}

// RefreshToken refresh user's access token
// @Summary Refresh Access Token
// @Description Refresh access token using valid refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.TokenRequest true "Payload Token Refresh"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "Unauthorized/Token Expired"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	requestId := requestid.GetRequestID(c)

	var request dto.TokenRequest
	if err := c.BodyParser(&request); err != nil {
		go logger.Log("SYSTEM", "ERROR", "Failed to parse request body: "+err.Error(), requestId)
		return customerrors.ErrBadRequest
	}

	res, err := h.authService.NewAccessToken(c.Context(), request.RefreshToken, requestId)
	if err != nil {
		go logger.Log("SYSTEM", "ERROR", "Failed to refresh token: "+err.Error(), requestId)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(
		response.StatusSuccess,
		"Token Refreshed Successfully",
		res,
		requestId,
	))
}
