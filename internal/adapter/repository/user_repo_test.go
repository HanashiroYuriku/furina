package repository_test

import (
	"be-ayaka/internal/adapter/repository"
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/port"
	"be-ayaka/internal/testingutils"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserRepoSuite struct {
	BaseRepoSuite // Mengembed base suite
	repo          port.UserRepository
}

func (s *UserRepoSuite) SetupSuite() {
	s.BaseRepoSuite.SetupSuite()
	s.repo = repository.NewUserRepo(s.DB)
}

func TestUserRepoSuite(t *testing.T) {
	suite.Run(t, &UserRepoSuite{
		BaseRepoSuite: NewBaseRepoSuite(&entity.User{}),
	})
}

func (s *UserRepoSuite) TestCreate() {
	ctx := context.Background()

	// success create
	s.Run("Success - Create User", func() {
		newUser := &entity.User{
			Username: "hanashiroyuriku",
			Email:    "yuriku@mail.com",
			Password: "hashedpassword",
		}
		newUser.ID = "USER-123"

		err := s.repo.Create(ctx, newUser)

		s.NoError(err)

		var res entity.User
		s.DB.First(&res, "id = ?", "USER-123")
		s.Equal("hanashiroyuriku", res.Username)
		s.Equal("yuriku@mail.com", res.Email)
	})

	// success create with tx
	s.Run("Success - Success Create with tx", func() {
		tx := s.DB.Begin()

		ctxWithTx := context.WithValue(context.Background(), repository.TxKey{}, tx)

		newUser := &entity.User{
			Username: "riku",
			Email: "riku@mail.com",
		}
		newUser.ID = "USER-1"

		err := s.repo.Create(ctxWithTx, newUser)
		s.NoError(err)

		tx.Rollback()

		var res entity.User
		errFind := s.DB.First(&res, "id = ?", "USER-1").Error
		s.Error(errFind)
		s.Equal(gorm.ErrRecordNotFound, errFind)
	})

	// failed duplicate
	s.Run("Failed - Duplicate email", func() {
		user1 := &entity.User{
			Username: "user1",
			Email:    "test@mail.com",
		}
		user1.ID = "U1"
		s.DB.Create(user1)

		user2 := &entity.User{
			Username: "user2",
			Email:    "test@mail.com",
		}
		user2.ID = "U2"

		err := s.repo.Create(ctx, user2)

		s.Error(err)
		s.Contains(err.Error(), "UNIQUE constraint failed")
	})
}

func (s *UserRepoSuite) TestFindByEmailUsername() {
	ctx := context.Background()

	dummy := &entity.User{
		Username: "riku",
		Email:    "riku@mail.com",
	}
	dummy.ID = "USER-001"
	s.DB.Create(dummy)

	// success find by email
	s.Run("Success - Find by email", func() {
		res, err := s.repo.FindByEmailUsername(ctx, "riku@mail.com")
		s.NoError(err)
		s.Equal("riku", res.Username)
	})

	// success find by username
	s.Run("Success - Find by username", func() {
		res, err := s.repo.FindByEmailUsername(ctx, "riku")
		s.NoError(err)
		s.Equal("riku@mail.com", res.Email)
	})

	// username or email not found
	s.Run("Failed - Not found", func() {
		res, err := s.repo.FindByEmailUsername(ctx, "unknown")
		s.Error(err)
		s.Nil(res)
		s.ErrorIs(err, customerrors.ErrInvalidCredentials)
	})
}

func (s *UserRepoSuite) TestUpdateRefreshToken() {
	ctx := context.Background()
	userId := "USER-001"

	s.DB.Create(&entity.User{
		BaseEntity: entity.BaseEntity{
			ID: userId,
		},
		Username: "riku",
	})

	// success update refresh token
	s.Run("Success - Success update refresh token", func() {
		newToken := "new token"
		err := s.repo.UpdateRefreshToken(ctx, userId, newToken)

		s.NoError(err)

		var res entity.User
		s.DB.First(&res, "id = ?", userId)
		s.NotNil(res.RefreshToken)
		s.Equal(newToken, *res.RefreshToken)
	})
}

func (s *UserRepoSuite) TestFindByID() {
	ctx := context.Background()

	dummy := &entity.User{
		Username: "riku",
		Email: "riku@mail.com",
	}
	dummy.ID = "USER-1"
	s.DB.Create(dummy)

	// success find user
	s.Run("Success - Success find user", func() {
		res, err := s.repo.FindByID(ctx, dummy.ID)
		s.NoError(err)
		s.Equal("riku", res.Username)
	})

	// id not found
	s.Run("Failed - ID User not found", func() {
		res, err := s.repo.FindByID(ctx, "unknown")
		s.Error(err)
		s.Nil(res)
		s.ErrorIs(err, customerrors.ErrDataNotFound)
	})
}

func (s *UserRepoSuite) TestFindByRefreshToken() {
	ctx := context.Background()

	dummy := &entity.User{
		Username: "riku",
		Email: "riku@mail.com",
		RefreshToken: testingutils.StringPtr("refreshToken"),
	}
	dummy.ID = "USER-1"
	s.DB.Create(dummy)

	// success find refresh token
	s.Run("Success - Success find refresh token", func() {
		res, err := s.repo.FindByRefreshToken(ctx, *dummy.RefreshToken)
		s.NoError(err)
		s.Equal("riku", res.Username)
	})

	// refresh token not found
	s.Run("Failed - Refresh token not found", func() {
		res, err := s.repo.FindByRefreshToken(ctx, "unknown")
		s.Error(err)
		s.Nil(res)
		s.ErrorIs(err, customerrors.ErrInvalidCredentials)
	})
}

func (s *UserRepoSuite) TestVerifyUser() {
	ctx := context.Background()
	id := "USER-1"
	s.DB.Create(&entity.User{
		BaseEntity: entity.BaseEntity{
			ID: id,
		},
		Username: "riku",
		Email: "riku@mail.com",
		IsVerified: false,
	})
	
	// success verify user
	s.Run("Success - Success Verify user", func ()  {
		err := s.repo.VerifUser(ctx, id)

		s.NoError(err)

		var res entity.User
		s.DB.First(&res, "id = ?", id)
		s.NotNil(res)
		s.Equal(res.IsVerified, true)
	})
}