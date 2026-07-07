package handlers

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/porfolio-roger-bass-api/internal/services"
)

// UploadHandler exposes image/video upload endpoints backed by Backblaze B2
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

var imageExts = map[string]string{
	".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
	".webp": "image/webp", ".gif": "image/gif", ".avif": "image/avif",
}

var videoExts = map[string]string{
	".mp4": "video/mp4", ".webm": "video/webm", ".mov": "video/quicktime",
	".avi": "video/x-msvideo", ".mkv": "video/x-matroska",
}

// UploadImage handles POST /api/admin/upload
func (h *UploadHandler) UploadImage(c *fiber.Ctx) error {
	return h.handleUpload(c, "image", 10<<20, "images")
}

// UploadVideo handles POST /api/admin/upload/video
func (h *UploadHandler) UploadVideo(c *fiber.Ctx) error {
	return h.handleUpload(c, "video", 100<<20, "videos")
}

func (h *UploadHandler) handleUpload(c *fiber.Ctx, kind string, maxSize int64, folder string) error {
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

	if file.Size > maxSize {
		limitMB := maxSize / (1 << 20)
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"error": fmt.Sprintf("Le fichier ne doit pas dépasser %d Mo", limitMB),
		})
	}

	contentType, valid := detectMediaType(file, kind)
	if !valid {
		if kind == "image" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Seules les images sont acceptées (JPEG, PNG, WebP, GIF…)",
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Seules les vidéos sont acceptées (MP4, WebM, MOV…)",
		})
	}

	url, err := h.b2.UploadFileWithContentType(file, folder, contentType)
	if err != nil {
		log.Printf("Erreur upload B2 (%s) : %v", kind, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Échec de l'upload vers Backblaze B2 : " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"url": url, "type": kind})
}

func detectMediaType(file *multipart.FileHeader, kind string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	headerCT := strings.ToLower(strings.TrimSpace(file.Header.Get("Content-Type")))

	var extMap map[string]string
	var prefix string
	if kind == "image" {
		extMap = imageExts
		prefix = "image/"
	} else {
		extMap = videoExts
		prefix = "video/"
	}

	// Trust extension first (browsers often send empty or wrong Content-Type)
	if ct, ok := extMap[ext]; ok {
		return ct, true
	}

	// Fall back to declared Content-Type
	if strings.HasPrefix(headerCT, prefix) {
		return headerCT, true
	}

	// Sniff first 512 bytes
	src, err := file.Open()
	if err != nil {
		return "", false
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, _ := io.ReadFull(src, buf)
	sniffed := http.DetectContentType(buf[:n])

	if strings.HasPrefix(sniffed, prefix) {
		return sniffed, true
	}

	// HEIC/HEIF sometimes detected as application/octet-stream
	if kind == "image" && (ext == ".heic" || ext == ".heif") {
		return "image/heic", true
	}

	return "", false
}
