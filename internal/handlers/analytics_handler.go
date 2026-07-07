package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/porfolio-roger-bass-api/internal/models"
	"gorm.io/gorm"
)

var countryNames = map[string]string{
	"FR": "France", "BE": "Belgique", "CH": "Suisse", "CA": "Canada",
	"US": "États-Unis", "GB": "Royaume-Uni", "DE": "Allemagne", "CD": "RD Congo",
	"CG": "Congo", "CI": "Côte d'Ivoire", "SN": "Sénégal", "CM": "Cameroun",
	"MA": "Maroc", "DZ": "Algérie", "TN": "Tunisie", "IT": "Italie",
	"ES": "Espagne", "PT": "Portugal", "NL": "Pays-Bas", "BR": "Brésil",
}

type AnalyticsHandler struct {
	db *gorm.DB
}

func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

type trackPayload struct {
	PagePath string `json:"page_path"`
	Referrer string `json:"referrer"`
}

// Track handles POST /api/analytics/track — records a page view (public)
func (h *AnalyticsHandler) Track(c *fiber.Ctx) error {
	var payload trackPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Corps invalide"})
	}

	pagePath := strings.TrimSpace(payload.PagePath)
	if pagePath == "" {
		pagePath = "/"
	}
	if len(pagePath) > 200 {
		pagePath = pagePath[:200]
	}

	code, country := detectCountry(c)

	view := models.PageView{
		PagePath:    pagePath,
		CountryCode: code,
		Country:     country,
		Referrer:    truncate(payload.Referrer, 500),
		UserAgent:   truncate(c.Get("User-Agent"), 500),
	}

	if err := h.db.Create(&view).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur enregistrement vue"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"ok": true})
}

// AdminStats handles GET /api/admin/analytics/stats
func (h *AnalyticsHandler) AdminStats(c *fiber.Ctx) error {
	var totalViews int64
	h.db.Model(&models.PageView{}).Count(&totalViews)

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfWeek := startOfDay.AddDate(0, 0, -int(now.Weekday()))
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var viewsToday, viewsWeek, viewsMonth int64
	h.db.Model(&models.PageView{}).Where("created_at >= ?", startOfDay).Count(&viewsToday)
	h.db.Model(&models.PageView{}).Where("created_at >= ?", startOfWeek).Count(&viewsWeek)
	h.db.Model(&models.PageView{}).Where("created_at >= ?", startOfMonth).Count(&viewsMonth)

	type countryStat struct {
		CountryCode string `json:"country_code"`
		Country     string `json:"country"`
		Count       int64  `json:"count"`
	}
	var byCountry []countryStat
	h.db.Model(&models.PageView{}).
		Select("country_code, country, count(*) as count").
		Group("country_code, country").
		Order("count desc").
		Limit(20).
		Scan(&byCountry)

	type pageStat struct {
		PagePath string `json:"page_path"`
		Count    int64  `json:"count"`
	}
	var byPage []pageStat
	h.db.Model(&models.PageView{}).
		Select("page_path, count(*) as count").
		Group("page_path").
		Order("count desc").
		Limit(10).
		Scan(&byPage)

	type dailyStat struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}
	var daily []dailyStat
	h.db.Model(&models.PageView{}).
		Select("to_char(created_at, 'YYYY-MM-DD') as date, count(*) as count").
		Where("created_at >= ?", now.AddDate(0, 0, -30)).
		Group("date").
		Order("date asc").
		Scan(&daily)

	return c.JSON(fiber.Map{
		"total_views":  totalViews,
		"views_today":  viewsToday,
		"views_week":   viewsWeek,
		"views_month":  viewsMonth,
		"by_country":   byCountry,
		"by_page":      byPage,
		"daily_views":  daily,
	})
}

func detectCountry(c *fiber.Ctx) (code, name string) {
	code = strings.ToUpper(strings.TrimSpace(c.Get("CF-IPCountry")))
	if code == "" {
		code = strings.ToUpper(strings.TrimSpace(c.Get("X-Country-Code")))
	}
	if code == "" || code == "XX" || code == "T1" {
		code = "Unknown"
		name = "Inconnu"
		return code, name
	}
	if n, ok := countryNames[code]; ok {
		name = n
	} else {
		name = code
	}
	return code, name
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
