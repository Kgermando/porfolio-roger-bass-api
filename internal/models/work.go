package models

import "gorm.io/gorm"

// Work represents a musical work or production
type Work struct {
	gorm.Model
	Title     string `json:"title" gorm:"not null;size:200"`
	Category  string `json:"category" gorm:"not null;size:50"` // performances | tutoriels | compositions
	Desc      string `json:"desc" gorm:"type:text"`
	Link      string `json:"link" gorm:"size:500"`
	IsActive  bool   `json:"is_active" gorm:"default:true"`
	SortOrder int    `json:"sort_order" gorm:"default:0"`
}
