package model

import (
	"time"
)

type AdminUser struct {
	ID           string    `json:"id" gorm:"column:id;type:uuid"`
	Username     string    `json:"username" gorm:"column:username;unique"`
	PasswordHash string    `json:"passwordHash" gorm:"column:password_hash;not null"`
	Role         string    `json:"role" gorm:"column:role;not null;default:'editor'"`
	CreatedAt    time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}
