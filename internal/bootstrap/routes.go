package bootstrap

import (
	"be-ayaka/config"
	_ "be-ayaka/docs"
	"be-ayaka/internal/delivery/http"
	"be-ayaka/internal/middleware"
	"be-ayaka/pkg/generate_id"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, cfg *config.Config, db *gorm.DB) {
	// middleware request
	app.Use(requestid.New(requestid.Config{
		Header:     fiber.HeaderXRequestID,
		ContextKey: "requestid",
		Generator: func() string {
			return generateid.GenerateID("REQUEST")
		},
	}))

	// === Swagger UI
	app.Get("/swagger/*", swagger.HandlerDefault)

	// === Health Check
	healthHandler := http.NewHealthCheckHandler(cfg, db)
	// --- health route
	app.Get("/health", healthHandler.Check)
	// ---

	// === Handler ===
	handler := BuildAllDependencies(db, cfg)

	// === api app group
	apiApp := app.Group("/api")

	// === version 1
	apiV1 := apiApp.Group("/v1")

	// === auth group
	authGroup := apiV1.Group("/auth")
	// --- register
	authGroup.Post("/register", handler.Auth.RegisterUser)
	// --- resend verif
	authGroup.Post("/resend-verification", handler.Auth.ResendVerification)
	// --- verify email
	authGroup.Get("/verify", handler.Auth.VerifyEmail)
	// --- login
	authGroup.Post("/login", handler.Auth.Login)
	// --- refresh token
	authGroup.Post("/refresh", handler.Auth.RefreshToken)

	// auth required
	authApp := apiApp.Group("/", middleware.RequireAuth(cfg))
	_ = authApp
}
