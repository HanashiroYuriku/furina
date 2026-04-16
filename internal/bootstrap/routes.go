package bootstrap

import (
	"be-ayaka/config"
	"be-ayaka/internal/delivery/http"
	"be-ayaka/internal/middleware"
	"be-ayaka/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, cfg *config.Config, db *gorm.DB) {
	// middleware request
	app.Use(requestid.New(requestid.Config{
		Header:     fiber.HeaderXRequestID,
		ContextKey: "request_id",
		Generator: func() string {
			return utils.GenerateID("REQUEST")
		},
	}))

	// Health Check
	healthHandler := http.NewHealthCheckHandler(cfg, db)
	// health route
	app.Get("/health", healthHandler.Check)
	// ---

	// === User Handler ===
	userHandler := BuildUserHandler(db)

	// --- Register User ---
	app.Post("/register", userHandler.RegisterUser)

	// version & auth
	apiApp := app.Group("/api/v1")
	auth := apiApp.Group("/", middleware.RequireAuth(cfg))
	_ = auth

	// User and Admin routes
	// userArea := auth.Group("/user", middleware.OnlyRole("user", "admin"))

	// user
	// userHandler := BuildUserHandler(db)

	// userArea.Get("/profile", userHandler.GetProfile)
}
