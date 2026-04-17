package port

import "be-ayaka/internal/core/entity"

type UserRepository interface {
	FindByID(id string) (*entity.User, error)
	FindByEmailUsername(value string) (*entity.UserResponse, error)
	Create(user *entity.User) error
	VerifUser(id string) error
}