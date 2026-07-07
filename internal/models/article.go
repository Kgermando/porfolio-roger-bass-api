package models

import "time"

// Article represents a biblical preaching / blog post
type Article struct {
	ID          uint       `json:"ID" gorm:"primaryKey"`
	CreatedAt   time.Time  `json:"CreatedAt"`
	UpdatedAt   time.Time  `json:"UpdatedAt"`
	DeletedAt   *time.Time `json:"DeletedAt,omitempty" gorm:"index"`
	Title       string     `json:"title" gorm:"not null;size:300"`
	Slug        string     `json:"slug" gorm:"uniqueIndex;not null;size:300"`
	Excerpt     string     `json:"excerpt" gorm:"type:text"`
	Content     string     `json:"content" gorm:"type:text;not null"`
	CoverImage  string     `json:"cover_image" gorm:"size:500"`
	Author      string     `json:"author" gorm:"size:200;default:'Roger Bass'"`
	IsPublished bool       `json:"is_published" gorm:"default:false;index"`
	SortOrder   int        `json:"sort_order" gorm:"default:0"`
	ViewCount   int        `json:"view_count" gorm:"default:0"`
	PublishedAt *time.Time `json:"published_at"`
}
