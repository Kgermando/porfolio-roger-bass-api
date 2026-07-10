package handlers

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/porfolio-roger-bass-api/internal/models"
	"gorm.io/gorm"
)

type ArticleHandler struct {
	db *gorm.DB
}

func NewArticleHandler(db *gorm.DB) *ArticleHandler {
	return &ArticleHandler{db: db}
}

// articleRequest is the JSON body for create/update (explicit fields avoid BodyParser issues)
type articleRequest struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Excerpt     string `json:"excerpt"`
	Content     string `json:"content"`
	CoverImage  string `json:"cover_image"`
	Author      string `json:"author"`
	IsPublished bool   `json:"is_published"`
	SortOrder   int    `json:"sort_order"`
}

// List handles GET /api/articles — published articles (public)
func (h *ArticleHandler) List(c *fiber.Ctx) error {
	var articles []models.Article
	if err := h.db.Where("is_published IS TRUE").
		Order("sort_order asc, published_at desc, created_at desc").
		Find(&articles).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur serveur"})
	}
	if articles == nil {
		articles = []models.Article{}
	}
	return c.JSON(articles)
}

// GetBySlug handles GET /api/articles/:slug — single published article + increment views
func (h *ArticleHandler) GetBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var article models.Article
	if err := h.db.Where("slug = ? AND is_published IS TRUE", slug).First(&article).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Article introuvable"})
	}
	h.db.Model(&article).UpdateColumn("view_count", gorm.Expr("view_count + 1"))
	article.ViewCount++
	return c.JSON(article)
}

// AdminList handles GET /api/admin/articles
func (h *ArticleHandler) AdminList(c *fiber.Ctx) error {
	var articles []models.Article
	if err := h.db.Order("sort_order asc, created_at desc").Find(&articles).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur serveur"})
	}
	return c.JSON(articles)
}

// Create handles POST /api/admin/articles
func (h *ArticleHandler) Create(c *fiber.Ctx) error {
	var input articleRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Corps invalide"})
	}
	if input.Title == "" || input.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Titre et contenu requis"})
	}

	article := models.Article{
		Title:       input.Title,
		Excerpt:     input.Excerpt,
		Content:     input.Content,
		CoverImage:  input.CoverImage,
		Author:      input.Author,
		IsPublished: input.IsPublished,
		SortOrder:   input.SortOrder,
	}
	if article.Slug == "" {
		article.Slug = slugify(input.Title)
	} else {
		article.Slug = slugify(input.Slug)
	}
	if article.Author == "" {
		article.Author = "Roger Bass"
	}
	if article.IsPublished {
		now := time.Now()
		article.PublishedAt = &now
	}
	if err := h.db.Create(&article).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Slug déjà utilisé ou erreur de création"})
	}
	return c.Status(fiber.StatusCreated).JSON(article)
}

// Update handles PUT /api/admin/articles/:id
func (h *ArticleHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID invalide"})
	}

	var article models.Article
	if err := h.db.First(&article, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Article introuvable"})
	}

	var input articleRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Corps invalide"})
	}

	wasPublished := article.IsPublished
	article.Title = input.Title
	article.Excerpt = input.Excerpt
	article.Content = input.Content
	article.CoverImage = input.CoverImage
	article.Author = input.Author
	article.IsPublished = input.IsPublished
	article.SortOrder = input.SortOrder
	if input.Author == "" {
		article.Author = "Roger Bass"
	}
	if input.Slug != "" {
		article.Slug = slugify(input.Slug)
	}
	if article.IsPublished && !wasPublished && article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
	}

	if err := h.db.Save(&article).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la mise à jour"})
	}
	return c.JSON(article)
}

// TogglePublish handles PUT /api/admin/articles/:id/publish — quick publish/unpublish
func (h *ArticleHandler) TogglePublish(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID invalide"})
	}

	var article models.Article
	if err := h.db.First(&article, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Article introuvable"})
	}

	article.IsPublished = !article.IsPublished
	if article.IsPublished && article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
	}

	if err := h.db.Save(&article).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la mise à jour"})
	}
	return c.JSON(article)
}

// Delete handles DELETE /api/admin/articles/:id
func (h *ArticleHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID invalide"})
	}
	if err := h.db.Delete(&models.Article{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la suppression"})
	}
	return c.JSON(fiber.Map{"message": "Article supprimé"})
}

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		} else if unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			b.WriteRune('-')
		} else {
			// strip accents roughly
			switch r {
			case 'à', 'â', 'ä', 'á', 'ã', 'å':
				b.WriteRune('a')
			case 'é', 'è', 'ê', 'ë':
				b.WriteRune('e')
			case 'î', 'ï', 'í':
				b.WriteRune('i')
			case 'ô', 'ö', 'ó', 'õ':
				b.WriteRune('o')
			case 'ù', 'û', 'ü', 'ú':
				b.WriteRune('u')
			case 'ç':
				b.WriteRune('c')
			default:
				b.WriteRune('-')
			}
		}
	}
	result := nonSlugChars.ReplaceAllString(b.String(), "-")
	result = strings.Trim(result, "-")
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	if result == "" {
		result = "article"
	}
	return result
}
