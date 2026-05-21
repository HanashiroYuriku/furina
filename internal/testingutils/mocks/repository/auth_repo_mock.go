package mocks

import (
	"be-ayaka/internal/core/entity"
	"context"
	"github.com/stretchr/testify/mock"
)

// Mock user verif repo
type MockUserVerificationRepo struct {
	mock.Mock
}

// func find by token
func (m *MockUserVerificationRepo) FindByToken(ctx context.Context, token string) (*entity.UserVerification, error) {
	args := m.Called(ctx, token)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.UserVerification), args.Error(1)
	}
	return nil, args.Error(1)
}

// func upsert
func (m *MockUserVerificationRepo) Upsert(ctx context.Context, data *entity.UserVerification) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

// func by email
func (m *MockUserVerificationRepo) FindByEmail(ctx context.Context, email string) (*entity.UserVerification, error) {
	args := m.Called(ctx, email)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.UserVerification), args.Error(1)
	}
	return nil, args.Error(1)
}
