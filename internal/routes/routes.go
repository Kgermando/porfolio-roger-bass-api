package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/porfolio-roger-bass-api/internal/handlers"
	"github.com/kgermando/porfolio-roger-bass-api/internal/middleware"
	"gorm.io/gorm"
)

// Setup registers all application routes
func Setup(app *fiber.App, db *gorm.DB) {
	contactHandler := handlers.NewContactHandler(db)
	eventHandler := handlers.NewEventHandler(db)
	workHandler := handlers.NewWorkHandler(db)
	galleryHandler := handlers.NewGalleryHandler(db)
	authHandler := handlers.NewAuthHandler(db)
	uploadHandler := handlers.NewUploadHandler()

	api := app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "Roger Bass API"})
	})

	// ── Public routes ────────────────────────────────
	api.Post("/contact", contactHandler.Create)
	api.Get("/events", eventHandler.List)
	api.Get("/works", workHandler.List)
	api.Get("/gallery", galleryHandler.List)

	// ── Auth routes ──────────────────────────────────
	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Get("/me", middleware.Protected(), authHandler.Me)

	// ── Admin routes (protected) ─────────────────────
	admin := api.Group("/admin", middleware.Protected())

	// Works CRUD
	admin.Get("/works", workHandler.AdminList)
	admin.Post("/works", workHandler.Create)
	admin.Put("/works/:id", workHandler.Update)
	admin.Delete("/works/:id", workHandler.Delete)

	// Events CRUD
	admin.Get("/events", eventHandler.AdminList)
	admin.Post("/events", eventHandler.Create)
	admin.Put("/events/:id", eventHandler.Update)
	admin.Delete("/events/:id", eventHandler.Delete)

	// Contacts (read-only + mark as read)
	admin.Get("/contacts", contactHandler.AdminList)
	admin.Put("/contacts/:id/read", contactHandler.MarkRead)
	admin.Delete("/contacts/:id", contactHandler.DeleteContact)

	// Gallery CRUD
	admin.Get("/gallery", galleryHandler.AdminList)
	admin.Post("/gallery", galleryHandler.Create)
	admin.Put("/gallery/:id", galleryHandler.Update)
	admin.Delete("/gallery/:id", galleryHandler.Delete)

	// Image upload → Backblaze B2
	admin.Post("/upload", uploadHandler.UploadImage)
}
