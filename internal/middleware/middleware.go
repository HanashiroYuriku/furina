package middleware

import (
	"strings"

	"be-ayaka/config"
	customjwt "be-ayaka/pkg/jwt"
	"be-ayaka/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestId, ok := c.Locals("request_id").(string)

		if !ok {
			requestId = c.Get("X-Request-ID", "unknown-request-id")
		}

		authHeader := c.Get("Authorization")

		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse(
				response.Unauthorized,
				"Authorization header is missing",
				requestId,
			))
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse(
				response.Unauthorized,
				"Invalid Token Format",
				requestId))
		}

		tokenString := parts[1]

		claims := &customjwt.CustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse(
				response.Unauthorized,
				"Token Invalid or Expired",
				requestId,
			))
		}

		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
