package models

import "gorm.io/gorm"

// GalleryPhoto represents a photo in the portfolio gallery
type GalleryPhoto struct {
	gorm.Model
	Src       string `json:"src" gorm:"not null;size:500"`
	Alt       string `json:"alt" gorm:"size:300"`
	Caption   string `json:"caption" gorm:"size:300"`
	IsActive  bool   `json:"is_active" gorm:"default:true"`
	SortOrder int    `json:"sort_order" gorm:"default:0"`
}
