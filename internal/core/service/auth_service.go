package service

import (
	"be-ayaka/config"
	"be-ayaka/internal/core/customerrors"
	"be-ayaka/internal/core/entity"
	"be-ayaka/internal/core/port"
	"be-ayaka/pkg/hash"
	"be-ayaka/pkg/jwt"
	"be-ayaka/pkg/logger"
	"be-ayaka/pkg/utils"
	"errors"
	"fmt"
	"net/url"
	"time"
)

type AuthService interface {
	Create(user *entity.UserRequest) error
	ResendVerification(email string) error
	VerifyUser(token string) error
	Login(emailUsername, password, requestId string) (*entity.LoginResponse, error)
}

type authServiceImpl struct {
	userRepo     port.UserRepository
	hashService  hash.HashService
	authRepo     port.UserVerificationRepository
	emailAdapter port.EmailSender
	config       *config.Config
}

func NewAuthService(repo port.UserRepository, hashService hash.HashService, authRepo port.UserVerificationRepository, emailAdapter port.EmailSender, cfg *config.Config) AuthService {
	return &authServiceImpl{
		userRepo:     repo,
		hashService:  hashService,
		authRepo:     authRepo,
		emailAdapter: emailAdapter,
		config:       cfg,
	}
}

func (s *authServiceImpl) Create(user *entity.UserRequest) error {
	passwordHash, err := s.hashService.HashPassword(user.Password)
	if err != nil {
		return errors.New("Failed to hash Password")
	}

	userModel := &entity.User{
		ID:       utils.GenerateID("USER"),
		Username: user.Username,
		Email:    user.Email,
		Password: passwordHash,
		Role:     "user",
	}

	if err := s.userRepo.Create(userModel); err != nil {
		return err
	}

	go s.generateAndSendVerif(userModel)
	return nil
}

func (s *authServiceImpl) ResendVerification(email string) error {
	data, err := s.authRepo.FindByEmail(email)
	if err != nil {
		return err
	}

	if time.Since(data.CreatedAt).Minutes() < 5 {
		return customerrors.ErrColldownActive
	}

	userData, err := s.userRepo.FindByID(data.ID)
	if err != nil {
		return err
	}

	if userData.IsVerified {
		return customerrors.ErrAccountAlreadyVerified
	}

	go s.generateAndSendVerif(userData)
	return nil
}

func (s *authServiceImpl) VerifyUser(token string) error {
	userVerif, err := s.authRepo.FindByToken(token)
	if err != nil {
		return err
	}

	if time.Now().After(userVerif.ExpiredAt) {
		return customerrors.ErrTokenExpired
	}

	user, err := s.userRepo.FindByID(userVerif.UserID)
	if err != nil {
		return err
	}

	if user.IsVerified {
		return customerrors.ErrAccountAlreadyVerified
	}

	return s.userRepo.VerifUser(userVerif.UserID)
}

func (s *authServiceImpl) Login(emailUsername, password, requestId string) (*entity.LoginResponse, error) {
	user, err := s.userRepo.FindByEmailUsername(emailUsername)
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

	if err := s.userRepo.UpdateRefreshToken(user.ID, tokens.RefreshToken); err != nil {
		return nil, err
	}

	go logger.Log(user.ID, "INFO", fmt.Sprintf("User %s logged in successfully", user.Username), requestId)

	data := &entity.LoginResponse{
		AccessToken: tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:   int64(s.config.JWT.Expired),
		User: entity.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	return data, nil
}


// internal function
func (s *authServiceImpl) generateAndSendVerif(user *entity.User) error {
	token := utils.GenerateID("TOKEN")

	verifData := &entity.UserVerification{
		ID:        utils.GenerateID("VERIF"),
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		ExpiredAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.authRepo.Upsert(verifData); err != nil {
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
