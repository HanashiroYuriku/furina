package entity

// User represents a user in the system.
type User struct {
	BaseEntity
	Username     string  `gorm:"unique;not null;type:varchar(20)"`
	Email        string  `gorm:"unique;not null;type:varchar(100)"`
	DisplayName  string  `gorm:"type:varchar(30)"`
	Password     string  `gorm:"not null;type:text"`
	Role         string  `gorm:"type:varchar(20);default:'user';check:role IN ('admin', 'user')"`
	IsVerified   bool    `gorm:"default:false"`
	RefreshToken *string `gorm:"type:text"`
}
