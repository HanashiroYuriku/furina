package bootstrap

import (
	adapterRepo "be-ayaka/internal/adapter/repository"
	"be-ayaka/internal/core/service"
	"be-ayaka/internal/delivery/http"
	"be-ayaka/pkg/hash"
	"be-ayaka/pkg/validator"

	"gorm.io/gorm"
)

func BuildUserHandler(db *gorm.DB) *http.UserHandler {
	// adapter
	userRepo := adapterRepo.NewUserRepoPostgres(db)
	// service
	hashService := hash.NewBcryptHash()
	userService := service.NewUserService(userRepo, hashService)
	// handler
	validator := validator.NewGoValidator(db)
	return http.NewUserHandler(userService, validator)
}