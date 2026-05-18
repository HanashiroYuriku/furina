package repository_test

import (
	"be-ayaka/internal/adapter/repository"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/port"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserVerifRepoSuite struct {
	BaseRepoSuite 
	repo          port.UserVerificationRepository
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

func (s *UserVerifRepoSuite) TestUpsert() {
	ctx := context.Background()

	// success upsert
	s.Run("Success - Upsert", func() {
		newData := &entity.UserVerification{
			ID: "VERIF-123",
			UserID: "USER-123",
			Email: "riku@mail.com",
			Token: "token",
		}

		err := s.repo.Upsert(ctx, newData)

		s.NoError(err)

		var res entity.UserVerification
		s.DB.First(&res, "id = ?", "VERIF-123")
		s.Equal("USER-123", res.UserID)
		s.Equal("riku@mail.com", res.Email)
		s.Equal("token", res.Token)
	})
}