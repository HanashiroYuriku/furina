package port

import (
	"be-ayaka/internal/core/entity"
	"context"
)

type UserVerificationRepository interface {
	Upsert(ctx context.Context, data *entity.UserVerification) error
	FindByToken(ctx context.Context, token string) (*entity.UserVerification, error)
	FindByEmail(ctx context.Context, email string) (*entity.UserVerification, error)
}