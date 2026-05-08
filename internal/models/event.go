package models

import (
	"time"

	"gorm.io/gorm"
)

// Event represents a concert or public appearance
type Event struct {
	gorm.Model
	Title       string    `json:"title" gorm:"not null;size:200"`
	Description string    `json:"description" gorm:"type:text"`
	Location    string    `json:"location" gorm:"size:200"`
	Date        time.Time `json:"date"`
	ImageURL    string    `json:"image_url" gorm:"size:500"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
}
