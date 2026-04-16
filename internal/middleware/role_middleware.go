package middleware

import (
	"be-ayaka/pkg/response"

	"github.com/gofiber/fiber/v2"
)

func OnlyRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestId, ok := c.Locals("request_id").(string)

		if !ok {
			requestId = c.Get("X-Request-ID", "unknown-request-id")
		}

		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			return c.Status(fiber.StatusForbidden).JSON(response.NewErrorResponse(
				response.Forbidden,
				"Access Denied: Role not found",
				requestId,
			))
		}

		isAllowed := false
		for _, role := range allowedRoles {
			if role == userRole {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(response.NewErrorResponse(
				response.Forbidden,
				"Access Denied",
				requestId,
			))
		}

		return c.Next()
	}
}
