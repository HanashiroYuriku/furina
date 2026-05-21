package entity

import "time"

type UserVerification struct {
	ID        string    `gorm:"primaryKey;type:varchar(50)"`
	UserID    string    `gorm:"uniqueIndex;not null;type:varchar(50)"`
	Email     string    `gorm:"unique;not null;type:varchar(255)"`
	Token     string    `gorm:"unique;not null;type:varchar(255)"`
	ExpiredAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}
