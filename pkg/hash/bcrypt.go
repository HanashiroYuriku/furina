package hash

import (
	"golang.org/x/crypto/bcrypt"
)

type HashService interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword string, password string) error
}

type bcryptHash struct{}

func NewBcryptHash() HashService {
	return &bcryptHash{}
}

// hash text
func (h *bcryptHash) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// compare text and hash
func (h *bcryptHash) ComparePassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}