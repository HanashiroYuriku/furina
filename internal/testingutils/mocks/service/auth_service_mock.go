package mocks

import (
	"be-ayaka/internal/delivery/http/dto"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Create(ctx context.Context, user *dto.UserRequest) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockAuthService) ResendVerification(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockAuthService) VerifyUser(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockAuthService) Login(ctx context.Context, emailUsername, password, requestId string) (*dto.LoginResponse, error) {
	args := m.Called(ctx, emailUsername, password, requestId)
	return args.Get(0).(*dto.LoginResponse), args.Error(1)
}

func (m *MockAuthService) NewAccessToken(ctx context.Context, refreshToken, requestId string) (*dto.TokenResponse, error) {
	args := m.Called(ctx, refreshToken, requestId)
	return args.Get(0).(*dto.TokenResponse), args.Error(1)
}
