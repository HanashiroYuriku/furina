package jwt

import (
	"time"

	"be-ayaka/config"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService interface {
	GenerateToken(cfg *config.Config, userID string, role string) (*TokenPair, error)
}

type jwtService struct{}

func NewTokenService() TokenService {
	return &jwtService{}
}

// CustomClaims defines the structure of JWT claims
type CustomClaims struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// GenerateToken generates a JWT token for a given user ID and role
func (s *jwtService) GenerateToken(cfg *config.Config, userID string, role string) (*TokenPair, error) {
	// access token
	accessExp := time.Now().Add(time.Duration(cfg.JWT.Expired) * time.Minute)
	accessClaims := &CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.App.Name,
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return nil, err
	}

	// refresh token
	refreshExp := time.Now().Add(time.Duration(cfg.JWT.RefreshExpired) * time.Minute)
	refreshClaims := &CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.App.Name,
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(cfg.JWT.RefreshSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
