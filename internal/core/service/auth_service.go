package service

import (
	"be-ayaka/config"
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/port"
	generateid "be-ayaka/pkg/generate_id"
	"be-ayaka/pkg/hash"
	"be-ayaka/pkg/jwt"
	"be-ayaka/pkg/logger"
	"context"
	"fmt"
	"net/url"
	"time"
)

type AuthService interface {
	Create(ctx context.Context, user *entity.UserRequest) error
	ResendVerification(ctx context.Context, email string) error
	VerifyUser(ctx context.Context, token string) error
	Login(ctx context.Context, emailUsername, password, requestId string) (*entity.LoginResponse, error)
	NewAccessToken(ctx context.Context, refreshToken, requestId string) (*entity.TokenResponse, error)
}

type authServiceImpl struct {
	userRepo     port.UserRepository
	hashService  hash.HashService
	authRepo     port.UserVerificationRepository
	emailAdapter port.EmailSender
	config       *config.Config
	txManager    port.TxManager
}

func NewAuthService(repo port.UserRepository, hashService hash.HashService, authRepo port.UserVerificationRepository, emailAdapter port.EmailSender, cfg *config.Config, txManager port.TxManager) AuthService {
	return &authServiceImpl{
		userRepo:     repo,
		hashService:  hashService,
		authRepo:     authRepo,
		emailAdapter: emailAdapter,
		config:       cfg,
		txManager:    txManager,
	}
}

func (s *authServiceImpl) Create(ctx context.Context, user *entity.UserRequest) error {
	passwordHash, err := s.hashService.HashPassword(user.Password)
	if err != nil {
		return customerrors.ErrFailHash
	}

	userModel := &entity.User{
		Username: user.Username,
		Email:    user.Email,
		Password: passwordHash,
		Role:     "user",
	}
	userModel.ID = generateid.GenerateID("USER")

	return s.txManager.WithTx(ctx, func(ctx context.Context) error {
		if err := s.userRepo.Create(ctx, userModel); err != nil {
			return err
		}

		if err := s.generateAndSendVerif(ctx, userModel); err != nil {
			return err
		}

		return nil
	})
}

func (s *authServiceImpl) ResendVerification(ctx context.Context, email string) error {
	data, err := s.authRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	if time.Since(data.CreatedAt).Minutes() < 5 {
		return customerrors.ErrCooldownActive
	}

	userData, err := s.userRepo.FindByID(ctx, data.UserID)
	if err != nil {
		return err
	}

	if userData.IsVerified {
		return customerrors.ErrAccountAlreadyVerified
	}

	if err := s.generateAndSendVerif(ctx, userData); err != nil {
		return err
	}

	return nil
}

func (s *authServiceImpl) VerifyUser(ctx context.Context, token string) error {
	userVerif, err := s.authRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	if time.Now().After(userVerif.ExpiredAt) {
		return customerrors.ErrTokenExpired
	}

	user, err := s.userRepo.FindByID(ctx, userVerif.UserID)
	if err != nil {
		return err
	}

	if user.IsVerified {
		return customerrors.ErrAccountAlreadyVerified
	}

	return s.userRepo.VerifUser(ctx, userVerif.UserID)
}

func (s *authServiceImpl) Login(ctx context.Context, emailUsername, password, requestId string) (*entity.LoginResponse, error) {
	user, err := s.userRepo.FindByEmailUsername(ctx, emailUsername)
	if err != nil {
		return nil, err
	}

	if !user.IsVerified {
		return nil, customerrors.ErrAccountInactive
	}

	if err := s.hashService.ComparePassword(user.Password, password); err != nil {
		return nil, customerrors.ErrInvalidPassword
	}

	tokens, err := jwt.GenerateToken(s.config, user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateRefreshToken(ctx, user.ID, tokens.RefreshToken); err != nil {
		return nil, err
	}

	go logger.Log(user.ID, "INFO", fmt.Sprintf("User %s logged in successfully", user.Username), requestId)

	data := &entity.LoginResponse{
		User: entity.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}
	data.AccessToken = tokens.AccessToken
	data.RefreshToken = tokens.RefreshToken
	data.ExpiresIn = int64(s.config.JWT.Expired)

	return data, nil
}

// internal function
func (s *authServiceImpl) generateAndSendVerif(ctx context.Context, user *entity.User) error {
	token := generateid.GenerateID("TOKEN")

	verifData := &entity.UserVerification{
		ID:        generateid.GenerateID("VERIF"),
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		ExpiredAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.authRepo.Upsert(ctx, verifData); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/verify-email?token=%s", s.config.Frontend.URL, url.QueryEscape(token))
	subject := "Email Verification"
	body := fmt.Sprintf("<h1>Hello %s</h1><p>Your account is ready. Please Verify to activate your account</p><h2>Your token: %s</h2><p>Valid until %s</p>",
		user.Username, url, verifData.ExpiredAt.Format(time.RFC1123),
	)
	go s.emailAdapter.SendEmail(user.Email, subject, body)

	return nil
}

func (s *authServiceImpl) NewAccessToken(ctx context.Context, refreshToken, requestId string) (*entity.TokenResponse, error) {
	if refreshToken == "" {
		return nil, customerrors.ErrUnauthorized
	}

	user, err := s.userRepo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	tokens, err := jwt.GenerateToken(s.config, user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateRefreshToken(ctx, user.ID, tokens.RefreshToken); err != nil {
		return nil, err
	}

	go logger.Log(user.ID, "INFO", fmt.Sprintf("User %s create new tokens successfully", user.Username), requestId)

	data := &entity.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    int64(s.config.JWT.Expired),
	}

	return data, nil
}
