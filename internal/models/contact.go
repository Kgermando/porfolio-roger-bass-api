package models

import "gorm.io/gorm"

// Contact represents a contact form submission
type Contact struct {
	gorm.Model
	Name    string `json:"name" gorm:"not null;size:100"`
	Email   string `json:"email" gorm:"not null;size:150"`
	Phone   string `json:"phone" gorm:"size:20"`
	Subject string `json:"subject" gorm:"size:200"`
	Message string `json:"message" gorm:"not null;type:text"`
	IsRead  bool   `json:"is_read" gorm:"default:false"`
}

// CreateContactInput is the validated input for creating a contact
type CreateContactInput struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}
