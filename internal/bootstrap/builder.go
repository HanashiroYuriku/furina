package bootstrap

import (
	"be-ayaka/config"
	"be-ayaka/internal/adapter/email"
	adapterRepo "be-ayaka/internal/adapter/repository"
	"be-ayaka/internal/core/service"
	"be-ayaka/internal/delivery/http"
	"be-ayaka/pkg/hash"
	"be-ayaka/pkg/validator"

	"gorm.io/gorm"
)

type Handlers struct {
	Auth *http.AuthHandler
}

func BuildAllDependencies(db *gorm.DB, cfg *config.Config) *Handlers {
	// === email adapter
	emailAdapter := email.NewSMTPAdapter(
		cfg.Email.Host,
		cfg.Email.Port,
		cfg.Email.User,
		cfg.Email.Pass,
	)

	// validator
	validator := validator.NewGoValidator(db)

	// === adapter
	// --- user
	userRepo := adapterRepo.NewUserRepoPostgres(db)
	// --- auth
	authRepo := adapterRepo.NewUserVerificationRepo(db)

	// === service
	// --- user
	hashService := hash.NewBcryptHash()
	// --- auth
	authService := service.NewAuthService(userRepo, hashService, authRepo, emailAdapter, cfg)

	return &Handlers{
		Auth: http.NewAuthHandler(authService, validator),
	}
}
