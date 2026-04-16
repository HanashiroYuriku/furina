package jwt

import  (
	"time"

	"be-ayaka/config"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims defines the structure of JWT claims
type CustomClaims struct {
	UserID string `json:"userId"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token for a given user ID and role
func GenerateToken(cfg *config.Config, userID string, role string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.JWT.Expired) * time.Hour)

	// create claims
	claims := &CustomClaims{
		UserID: userID,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Issuer: cfg.App.Name,
		},
	}

	// create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign token with secret
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}