package requestid

import "github.com/gofiber/fiber/v2"

func GetRequestID(c *fiber.Ctx) string {
	requestId, ok := c.Locals("requestid").(string)
	if !ok || requestId == "" {
		return c.Get("X-Request-ID", "unknown-request-id")
	}
	return requestId
}
