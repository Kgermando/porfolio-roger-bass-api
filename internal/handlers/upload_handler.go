package handlers

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/porfolio-roger-bass-api/internal/services"
)

// UploadHandler exposes a single image-upload endpoint backed by Backblaze B2
type UploadHandler struct {
	b2 *services.BackblazeService
}

// NewUploadHandler initialises the handler. When B2 credentials are missing the
// handler is still created but every upload request will return 503.
func NewUploadHandler() *UploadHandler {
	b2, err := services.NewBackblazeService()
	if err != nil {
		log.Printf("Backblaze non configuré : %v", err)
		return &UploadHandler{}
	}
	return &UploadHandler{b2: b2}
}

// UploadImage handles POST /api/admin/upload
// Accepts a multipart field named "file", validates it is an image, uploads it
// to Backblaze B2 and returns { "url": "..." }.
func (h *UploadHandler) UploadImage(c *fiber.Ctx) error {
	if h.b2 == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Service de stockage non configuré (vérifiez les variables BACKBLAZE_*)",
		})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Champ 'file' manquant dans le formulaire multipart",
		})
	}

	// Validate MIME type
	ct := file.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Seules les images sont acceptées (image/jpeg, image/png, image/webp…)",
		})
	}

	// Limit: 10 MB
	const maxSize = 10 << 20
	if file.Size > maxSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"error": "L'image ne doit pas dépasser 10 Mo",
		})
	}

	url, err := h.b2.UploadImage(file)
	if err != nil {
		log.Printf("Erreur upload B2 : %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Échec de l'upload vers Backblaze B2",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"url": url})
}
