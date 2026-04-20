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

	// === Health Check
	healthHandler := http.NewHealthCheckHandler(cfg, db)
	// --- health route
	app.Get("/health", healthHandler.Check)
	// ---

	// === Handler ===
	handler := BuildAllDependencies(db, cfg)

	// === auth
	authGroup := app.Group("/auth")
	// --- register
	authGroup.Post("/register", handler.Auth.RegisterUser)
	// --- resend verif
	authGroup.Post("/resend-verification", handler.Auth.ResendVerification)
	// --- verify email
	authGroup.Get("/verify", handler.Auth.VerifyEmail)
	// --- login
	authGroup.Post("/login", handler.Auth.Login)

	// version & auth require
	apiApp := app.Group("/api/v1")
	auth := apiApp.Group("/", middleware.RequireAuth(cfg))
	_ = auth
}
