package port

import "be-ayaka/internal/core/entity"

type UserVerificationRepository interface {
	Upsert(data *entity.UserVerification) error
	FindByToken(token string) (*entity.UserVerification, error)
	FindByEmail(email string) (*entity.UserVerification, error)
}