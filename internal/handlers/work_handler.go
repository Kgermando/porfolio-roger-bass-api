package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/porfolio-roger-bass-api/internal/models"
	"gorm.io/gorm"
)

// WorkHandler handles work/portfolio requests
type WorkHandler struct {
	db *gorm.DB
}

// NewWorkHandler creates a new WorkHandler
func NewWorkHandler(db *gorm.DB) *WorkHandler {
	return &WorkHandler{db: db}
}

// List handles GET /api/works — paginated public list of active works
// Query params: category, page (default 1), limit (default 6)
func (h *WorkHandler) List(c *fiber.Ctx) error {
	category := c.Query("category", "")

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "6"))
	if err != nil || limit < 1 || limit > 50 {
		limit = 6
	}
	offset := (page - 1) * limit

	base := h.db.Model(&models.Work{}).Where("is_active = ?", true)
	if category != "" && category != "all" {
		base = base.Where("category = ?", category)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Impossible de compter les œuvres"})
	}

	var works []models.Work
	if err := base.Order("sort_order ASC, created_at ASC").Limit(limit).Offset(offset).Find(&works).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Impossible de récupérer les œuvres"})
	}
	if works == nil {
		works = []models.Work{}
	}

	pages := int(total) / limit
	if int(total)%limit != 0 {
		pages++
	}

	return c.JSON(fiber.Map{
		"data":  works,
		"total": total,
		"page":  page,
		"limit": limit,
		"pages": pages,
	})
}

// AdminList handles GET /api/admin/works — all works including inactive
func (h *WorkHandler) AdminList(c *fiber.Ctx) error {
	var works []models.Work
	if err := h.db.Order("sort_order ASC, created_at DESC").Find(&works).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Impossible de récupérer les œuvres"})
	}
	return c.JSON(works)
}

// Create handles POST /api/admin/works
func (h *WorkHandler) Create(c *fiber.Ctx) error {
	var work models.Work
	if err := c.BodyParser(&work); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Corps invalide"})
	}
	if work.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Le titre est requis"})
	}
	if err := h.db.Create(&work).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la création"})
	}
	return c.Status(fiber.StatusCreated).JSON(work)
}

// Update handles PUT /api/admin/works/:id
func (h *WorkHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID invalide"})
	}

	var work models.Work
	if err := h.db.First(&work, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Œuvre introuvable"})
	}

	var input models.Work
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Corps invalide"})
	}

	work.Title = input.Title
	work.Category = input.Category
	work.Desc = input.Desc
	work.Link = input.Link
	work.IsActive = input.IsActive
	work.SortOrder = input.SortOrder

	if err := h.db.Save(&work).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la mise à jour"})
	}
	return c.JSON(work)
}

// Delete handles DELETE /api/admin/works/:id
func (h *WorkHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID invalide"})
	}

	if err := h.db.Delete(&models.Work{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la suppression"})
	}
	return c.JSON(fiber.Map{"message": "Œuvre supprimée"})
}
