package models

import "gorm.io/gorm"

// Admin is the back-office user (Roger Bass)
type Admin struct {
	gorm.Model
	Username string `json:"username" gorm:"uniqueIndex;not null;size:100"`
	Password string `json:"-" gorm:"not null"` // bcrypt hash — never serialised
	FullName string `json:"full_name" gorm:"size:200"`
}

// LoginInput is the body expected by POST /api/auth/login
type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
