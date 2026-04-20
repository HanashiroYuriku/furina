package port

import "be-ayaka/internal/core/entity"

type UserRepository interface {
	FindByID(id string) (*entity.User, error)
	FindByEmailUsername(value string) (*entity.User, error)
	Create(user *entity.User) error
	VerifUser(id string) error
	UpdateRefreshToken(id string, token string) error
}