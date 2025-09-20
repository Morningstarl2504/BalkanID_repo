package models

import "time"

type AuditLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     *uint     `json:"user_id" gorm:"index"`
	Action     string    `json:"action" gorm:"not null"`
	Resource   string    `json:"resource" gorm:"not null"`
	ResourceID *uint     `json:"resource_id" gorm:"index"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Details    string    `json:"details" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at"`

	User *User `json:"user" gorm:"foreignKey:UserID"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}