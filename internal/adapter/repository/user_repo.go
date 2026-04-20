package repository

import (
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/port"
	"errors"

	"gorm.io/gorm"
)

type userRepoPostgres struct {
	db *gorm.DB
}

func NewUserRepoPostgres(db *gorm.DB) port.UserRepository {
	return &userRepoPostgres{
		db: db,
	}
}

func (r *userRepoPostgres) FindByID(id string) (*entity.User, error) {
	var user entity.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.ErrDataNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepoPostgres) FindByEmailUsername(value string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("email = ? OR username = ?", value, value).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.ErrInvalidCredentials
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepoPostgres) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *userRepoPostgres) VerifUser(id string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Update("is_verified", true).Error
}

func (r *userRepoPostgres) UpdateRefreshToken(id string, token string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Update("refresh_token", token).Error
}