package mocks

import (
	"be-ayaka/internal/core/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

// mock user repo
type MockUserRepo struct {
	mock.Mock
}

// func find by id
func (m *MockUserRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// func verif user
func (m *MockUserRepo) VerifUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepo) FindByEmailUsername(ctx context.Context, value string) (*entity.User, error) {
	args := m.Called(ctx, value)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) UpdateRefreshToken(ctx context.Context, id string, token string) error {
	args := m.Called(ctx, id, token)
	return args.Error(0)
}

func (m *MockUserRepo) FindByRefreshToken(ctx context.Context, token string) (*entity.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.User), args.Error(1)
	}
	return nil, args.Error(1)
}
