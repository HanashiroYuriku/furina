package middleware

import (
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/pkg/response"
	"be-ayaka/pkg/utils"
	"errors"

	"github.com/gofiber/fiber/v2"
)

func GlobalErrorHandler(c *fiber.Ctx, err error) error {
	requestId := utils.GetRequestID(c)

	var valErr *customerrors.ValidationError

	if errors.As(err, &valErr) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(
			response.NewErrorFieldResponse(
				response.UnprocessableEntity,
				valErr.Error(),
				valErr.Detail,
				requestId,
			),
		)
	}

	code := fiber.StatusInternalServerError
	statusString := response.InternalServerError
	message := "Internal Server Error"

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
		message = fiberErr.Message
	}

	switch {
	case errors.Is(err, customerrors.ErrDataNotFound):
		code = fiber.StatusNotFound
		statusString = response.DataNotFound
		message = err.Error()

	case errors.Is(err, customerrors.ErrInvalidPassword),
		errors.Is(err, customerrors.ErrTokenExpired):
		code = fiber.StatusUnauthorized
		statusString = response.Unauthorized
		message = err.Error()

	case errors.Is(err, customerrors.ErrDataConflict):
		code = fiber.StatusConflict
		statusString = response.DataConflict
		message = err.Error()

	case errors.Is(err, customerrors.ErrBadRequest):
		code = fiber.StatusBadRequest
		statusString = response.BadRequest
		message = err.Error()

	case errors.Is(err, customerrors.ErrColldownActive):
		code = fiber.StatusTooManyRequests
		statusString = response.TooManyRequests
		message = err.Error()

	case errors.Is(err, customerrors.ErrTokenExpired):
		code = fiber.StatusUnauthorized
		statusString = response.Unauthorized
		message = err.Error()

	case errors.Is(err, customerrors.ErrInvalidCredentials):
		code = fiber.StatusUnauthorized
		statusString = response.Unauthorized
		message = err.Error()

	case errors.Is(err, customerrors.ErrAccountInactive):
		code = fiber.StatusUnauthorized
		statusString = response.Unauthorized
		message = err.Error()

	case errors.Is(err, customerrors.ErrAccountAlreadyVerified):
		code = fiber.StatusConflict
		statusString = response.DataConflict
		message = err.Error()
	}

	return c.Status(code).JSON(response.NewErrorResponse(
		statusString,
		message,
		requestId,
	))
}
