package entity

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system.
type User struct {
	ID           string  `gorm:"primaryKey;type:varchar(50)"`
	Username     string  `gorm:"unique;not null;type:varchar(20)"`
	Email        string  `gorm:"unique;not null;type:varchar(100)"`
	Password     string  `gorm:"not null;type:varchar(255)"`
	Role         string  `gorm:"type:varchar(20);default:'user';check:role IN ('admin', 'user')"`
	IsVerified   bool    `gorm:"default:false"`
	RefreshToken *string `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// user struct for api response
type UserResponse struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	IsVerified bool   `json:"isVerified,omitempty"`
}

// user struct for api request
type UserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20,unique=user->username,whitespace,username"`
	Email    string `json:"email" validate:"required,max=100,email,unique=user->email,whitespace"`
	Password string `json:"password" validate:"required,complexpassword"`
}

type LoginResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	ExpiresIn    int64        `json:"expiresIn"`
	User         UserResponse `json:"user"`
}

type LoginRequest struct {
	EmailUsername string `json:"emailUsername" validate:"required"`
	Password      string `json:"password" validate:"required"`
}