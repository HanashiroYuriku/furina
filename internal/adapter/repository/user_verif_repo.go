package repository

import (
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/port"
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userVerificationRepo struct {
	db *gorm.DB
}

func NewUserVerificationRepo(db *gorm.DB) port.UserVerificationRepository {
	return &userVerificationRepo{
		db: db,
	}
}

func (r *userVerificationRepo) Upsert(ctx context.Context, data *entity.UserVerification) error {
	db := ExtractTx(ctx, r.db)

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"token", "expired_at", "created_at"}),
	}).Create(data).Error
}

func (r *userVerificationRepo) FindByToken(ctx context.Context, token string) (*entity.UserVerification, error) {
	var userVerif entity.UserVerification
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&userVerif).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, customerrors.ErrDataNotFound
		}
		return nil, err
	}
	return &userVerif, nil
}

func (r *userVerificationRepo) FindByEmail(ctx context.Context, email string) (*entity.UserVerification, error) {
	var user entity.UserVerification
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.ErrDataNotFound
		}
		return nil, err
	}
	return &user, nil
}
