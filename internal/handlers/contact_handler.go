package handlers

import (
	"log"
	"net/mail"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/porfolio-roger-bass-api/internal/models"
	"github.com/kgermando/porfolio-roger-bass-api/internal/services"
	"gorm.io/gorm"
)

// ContactHandler handles contact form requests
type ContactHandler struct {
	db    *gorm.DB
	email *services.EmailService
}

// NewContactHandler creates a new ContactHandler
func NewContactHandler(db *gorm.DB) *ContactHandler {
	return &ContactHandler{
		db:    db,
		email: services.NewEmailService(),
	}
}

// Create handles POST /api/contact
func (h *ContactHandler) Create(c *fiber.Ctx) error {
	var input models.CreateContactInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Corps de requête invalide",
		})
	}

	// Sanitize inputs
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Phone = strings.TrimSpace(input.Phone)
	input.Subject = strings.TrimSpace(input.Subject)
	input.Message = strings.TrimSpace(input.Message)

	// Validate required fields
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Le nom est requis"})
	}
	if input.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "L'email est requis"})
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Adresse email invalide"})
	}
	if input.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Le message est requis"})
	}

	// Length limits
	if len(input.Name) > 100 || len(input.Email) > 150 || len(input.Message) > 2000 || len(input.Subject) > 200 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Données trop longues"})
	}

	contact := models.Contact{
		Name:    input.Name,
		Email:   input.Email,
		Phone:   input.Phone,
		Subject: input.Subject,
		Message: input.Message,
	}

	if err := h.db.Create(&contact).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Échec de l'envoi du message",
		})
	}

	// Notify admin by email (non-blocking, failure is logged only)
	go func() {
		if err := h.email.SendContactNotification(
			contact.Name, contact.Email, contact.Subject, contact.Message, contact.Phone,
		); err != nil {
			log.Printf("Notification email échouée : %v", err)
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Message envoyé avec succès",
		"id":      contact.ID,
	})
}

// AdminList handles GET /api/admin/contacts — returns all messages, newest first
func (h *ContactHandler) AdminList(c *fiber.Ctx) error {
	var contacts []models.Contact
	if err := h.db.Order("created_at DESC").Find(&contacts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Impossible de récupérer les messages"})
	}
	return c.JSON(contacts)
}

// MarkRead handles PUT /api/admin/contacts/:id/read — marks a message as read
func (h *ContactHandler) MarkRead(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.db.Model(&models.Contact{}).Where("id = ?", id).Update("is_read", true).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Impossible de mettre à jour"})
	}
	return c.JSON(fiber.Map{"message": "Message marqué comme lu"})
}

// DeleteContact handles DELETE /api/admin/contacts/:id
func (h *ContactHandler) DeleteContact(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.db.Delete(&models.Contact{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la suppression"})
	}
	return c.JSON(fiber.Map{"message": "Message supprimé"})
}
