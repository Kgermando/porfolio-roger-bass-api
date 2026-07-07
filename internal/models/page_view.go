package models

import "gorm.io/gorm"

// PageView records a single page visit for analytics
type PageView struct {
	gorm.Model
	PagePath    string `json:"page_path" gorm:"size:200;index"`
	CountryCode string `json:"country_code" gorm:"size:10;index"`
	Country     string `json:"country" gorm:"size:100"`
	Referrer    string `json:"referrer" gorm:"size:500"`
	UserAgent   string `json:"user_agent" gorm:"size:500"`
}
