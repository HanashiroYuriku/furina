package entity

import "time"

type AuditLog struct {
	ID          string    `gorm:"primaryKey;type:varchar(50)"`
	UserID      string    `gorm:"not null;type:varchar(50)"`
	Action      string    `gorm:"not null;type:varchar(20);check:action IN ('CREATE', 'UPDATE', 'DELETE')"`
	TableSource string    `gorm:"not null;type:varchar(50)"`
	RecordID    string    `gorm:"not null;type:varchar(50)"`
	OldData     *string   `gorm:"type:text"` // string JSON format
	NewData     *string   `gorm:"type:text"` //string JSON format
	CreatedAt   time.Time `gorm:"not null"`
}

func (AuditLog) TableName() string { return "sys_audit_logs" }
