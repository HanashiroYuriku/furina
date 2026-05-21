package repository

import (
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/port"
	"context"
	"errors"

	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) port.UserRepository {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.ErrDataNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByEmailUsername(ctx context.Context, value string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ? OR username = ?", value, value).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.ErrInvalidCredentials
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Create(ctx context.Context, user *entity.User) error {
	db := ExtractTx(ctx, r.db)
	return db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) VerifUser(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("is_verified", true).Error
}

func (r *userRepo) UpdateRefreshToken(ctx context.Context, id string, token string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("refresh_token", token).Error
}

func (r *userRepo) FindByRefreshToken(ctx context.Context, token string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("refresh_token = ?", token).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.ErrInvalidCredentials
		}
		return nil, err
	}
	return &user, nil
}
