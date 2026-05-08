package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/porfolio-roger-bass-api/internal/models"
	"gorm.io/gorm"
)

// EventHandler handles event requests
type EventHandler struct {
	db *gorm.DB
}

// NewEventHandler creates a new EventHandler
func NewEventHandler(db *gorm.DB) *EventHandler {
	return &EventHandler{db: db}
}

// List handles GET /api/events — upcoming active events (public)
func (h *EventHandler) List(c *fiber.Ctx) error {
	var events []models.Event
	now := time.Now()

	if err := h.db.
		Where("date >= ? AND is_active = ?", now, true).
		Order("date ASC").
		Find(&events).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Impossible de récupérer les événements",
		})
	}
	return c.JSON(events)
}

// AdminList handles GET /api/admin/events — all events including past/inactive
func (h *EventHandler) AdminList(c *fiber.Ctx) error {
	var events []models.Event
	if err := h.db.Order("date DESC").Find(&events).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Impossible de récupérer les événements"})
	}
	return c.JSON(events)
}

// Create handles POST /api/admin/events
func (h *EventHandler) Create(c *fiber.Ctx) error {
	var event models.Event
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Corps invalide"})
	}
	if event.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Le titre est requis"})
	}
	if err := h.db.Create(&event).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la création"})
	}
	return c.Status(fiber.StatusCreated).JSON(event)
}

// Update handles PUT /api/admin/events/:id
func (h *EventHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID invalide"})
	}

	var event models.Event
	if err := h.db.First(&event, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Événement introuvable"})
	}

	var input models.Event
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Corps invalide"})
	}

	event.Title = input.Title
	event.Description = input.Description
	event.Location = input.Location
	event.Date = input.Date
	event.ImageURL = input.ImageURL
	event.IsActive = input.IsActive

	if err := h.db.Save(&event).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la mise à jour"})
	}
	return c.JSON(event)
}

// Delete handles DELETE /api/admin/events/:id
func (h *EventHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID invalide"})
	}

	if err := h.db.Delete(&models.Event{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la suppression"})
	}
	return c.JSON(fiber.Map{"message": "Événement supprimé"})
}
