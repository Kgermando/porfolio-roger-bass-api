package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/porfolio-roger-bass-api/internal/models"
	"gorm.io/gorm"
)

type GalleryHandler struct {
	db *gorm.DB
}

func NewGalleryHandler(db *gorm.DB) *GalleryHandler {
	return &GalleryHandler{db: db}
}

// List returns active photos ordered by sort_order (public)
func (h *GalleryHandler) List(c *fiber.Ctx) error {
	var photos []models.GalleryPhoto
	if err := h.db.Where("is_active = ?", true).Order("sort_order asc, created_at asc").Find(&photos).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur serveur"})
	}
	return c.JSON(photos)
}

// AdminList returns all photos including inactive (protected)
func (h *GalleryHandler) AdminList(c *fiber.Ctx) error {
	var photos []models.GalleryPhoto
	if err := h.db.Order("sort_order asc").Find(&photos).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur serveur"})
	}
	return c.JSON(photos)
}

// Create adds a new gallery photo (protected)
func (h *GalleryHandler) Create(c *fiber.Ctx) error {
	photo := new(models.GalleryPhoto)
	if err := c.BodyParser(photo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Corps de requête invalide"})
	}
	if photo.Src == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Le champ src est requis"})
	}
	if err := h.db.Create(photo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la création"})
	}
	return c.Status(fiber.StatusCreated).JSON(photo)
}

// Update modifies an existing gallery photo (protected)
func (h *GalleryHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var photo models.GalleryPhoto
	if err := h.db.First(&photo, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Photo introuvable"})
	}
	if err := c.BodyParser(&photo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Corps de requête invalide"})
	}
	if err := h.db.Save(&photo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la mise à jour"})
	}
	return c.JSON(photo)
}

// Delete removes a gallery photo (protected)
func (h *GalleryHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.db.Delete(&models.GalleryPhoto{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la suppression"})
	}
	return c.JSON(fiber.Map{"message": "Photo supprimée"})
}
