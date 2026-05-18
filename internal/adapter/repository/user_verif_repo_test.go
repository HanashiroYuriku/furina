package repository_test

import (
	"be-ayaka/internal/adapter/repository"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/port"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserVerifRepoSuite struct {
	BaseRepoSuite
	repo port.UserVerificationRepository
}

func (s *UserVerifRepoSuite) SetupSuite() {
	s.BaseRepoSuite.SetupSuite()
	s.repo = repository.NewUserVerificationRepo(s.DB)
}

func TestUserVerifRepoSuite(t *testing.T) {
	suite.Run(t, &UserVerifRepoSuite{
		BaseRepoSuite: NewBaseRepoSuite(&entity.UserVerification{}),
	})
}

// =============================================================================
// UPSERT TESTS
// =============================================================================
// 1. success scenario
func (s *UserVerifRepoSuite) TestUpsert_Success() {
	ctx := context.Background()

	newData := &entity.UserVerification{
		ID:     "VERIF-123",
		UserID: "USER-123",
		Email:  "riku@mail.com",
		Token:  "token",
	}

	err := s.repo.Upsert(ctx, newData)
	s.NoError(err)

	var res entity.UserVerification
	s.DB.First(&res, "id = ?", "VERIF-123")
	s.Equal("USER-123", res.UserID)
	s.Equal("riku@mail.com", res.Email)
	s.Equal("token", res.Token)
}

// 2. success scenario with tx
func (s *UserVerifRepoSuite) TestUpsert_SuccessWithTx() {
	tx := s.DB.Begin()
	ctxWithTx := context.WithValue(context.Background(), repository.TxKey{}, tx)

	newData := &entity.UserVerification{
		ID:     "VERIF-123",
		UserID: "USER-123",
		Email:  "riku@mail.com",
		Token:  "token",
	}

	err := s.repo.Upsert(ctxWithTx, newData)
	s.NoError(err)

	// Batalkan transaksi untuk memastikan data ditarik kembali dari DB RAM
	tx.Rollback()

	// Pembuktian isolasi data: data harus bersih dan mengembalikan ErrRecordNotFound
	var res entity.UserVerification
	errFind := s.DB.First(&res, "id = ?", "VERIF-123").Error
	s.Error(errFind)
	s.ErrorIs(errFind, gorm.ErrRecordNotFound)
}

// 3. failed scenario
func (s *UserVerifRepoSuite) TestUpsert_Failed_ContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	newData := &entity.UserVerification{
		ID:     "VERIF-123",
		UserID: "USER-123",
		Email:  "riku@mail.com",
		Token:  "token",
	}

	err := s.repo.Upsert(ctx, newData)

	// Asersi bahwa harus mengembalikan error (context canceled)
	s.Error(err)
}
// =============================================================================
// UPSERT TESTS
// =============================================================================