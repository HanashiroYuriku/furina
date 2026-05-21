package port

import (
	"be-ayaka/internal/core/entity"
	"context"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmailUsername(ctx context.Context, value string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	VerifUser(ctx context.Context, id string) error
	UpdateRefreshToken(ctx context.Context, id string, token string) error
	FindByRefreshToken(ctx context.Context, token string) (*entity.User, error)
}
