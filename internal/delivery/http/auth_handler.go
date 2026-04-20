package http

import (
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/service"
	"be-ayaka/pkg/logger"
	"be-ayaka/pkg/response"
	"be-ayaka/pkg/utils"
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

func (h *AuthHandler) RegisterUser(c *fiber.Ctx) error {
	requestId := utils.GetRequestID(c)

	var request entity.UserRequest

	if err := c.BodyParser(&request); err != nil {
		go logger.Log("SYSTEM", "ERROR", "Failed to parse request body: "+err.Error(), requestId)
		return customerrors.ErrBadRequest
	}

	if err := h.validator.Validate(c.Context(), &request); err != nil {
		go logger.Log("SYSTEM", "WARN", "Validation failed: "+err.Error(), requestId)
		return customerrors.NewValidationError(err.Error())
	}

	if err := h.authService.Create(&request); err != nil {
		go logger.Log("SYSTEM", "ERROR", "Internal Server Error: "+err.Error(), requestId)
		return err
	}

	go logger.Log("SYSTEM", "INFO", fmt.Sprintf("User %s created successfully", request.Username), requestId)
	return c.Status(fiber.StatusOK).JSON(
		response.NewSuccessResponse(
			response.StatusSuccess,
			"Success Create User",
			nil,
			requestId,
		),
	)
}

func (h *AuthHandler) ResendVerification(c *fiber.Ctx) error {
	requestId := utils.GetRequestID(c)

	var email entity.UserVerificationRequest
	if err := c.BodyParser(&email); err != nil {
		go logger.Log("SYSTEM", "ERROR", "Failed to parse request body: "+err.Error(), requestId)
		return customerrors.ErrBadRequest
	}

	if err := h.validator.Validate(c.Context(), &email); err != nil {
		go logger.Log("SYSTEM", "WARN", "Validation failed: "+err.Error(), requestId)
		return customerrors.NewValidationError(err.Error())
	}

	if err := h.authService.ResendVerification(email.Email); err != nil {
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

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	requestId := utils.GetRequestID(c)

	token := c.Query("token")
	if token == "" {
		go logger.Log("SYSTEM", "WARN", "Token query parameter is missing", requestId)
		return customerrors.ErrBadRequest
	}

	if err := h.authService.VerifyUser(token); err != nil {
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

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	requestId := utils.GetRequestID(c)

	var request entity.LoginRequest
	if err := c.BodyParser(&request); err != nil {
		go logger.Log("SYSTEM", "ERROR", "Failed to parse request body: "+err.Error(), requestId)
		return customerrors.ErrBadRequest
	}

	if err := h.validator.Validate(c.Context(), &request); err != nil {
		go logger.Log("SYSTEM", "WARN", "Validation failed: "+err.Error(), requestId)
		return err
	}

	res, err := h.authService.Login(request.EmailUsername, request.Password, requestId)
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
