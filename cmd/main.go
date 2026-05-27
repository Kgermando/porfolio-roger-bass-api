package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/kgermando/porfolio-roger-bass-api/internal/config"
	"github.com/kgermando/porfolio-roger-bass-api/internal/database"
	"github.com/kgermando/porfolio-roger-bass-api/internal/routes"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Échec de la connexion à la base de données:", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("Échec de la migration:", err)
	}

	database.Seed(db)

	app := fiber.New(fiber.Config{
		AppName:      "Roger Bass API v1.0",
		ErrorHandler: errorHandler,
		BodyLimit:    10 * 1024 * 1024, // 10 MB — aligns with the upload handler limit
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	routes.Setup(app, db)

	log.Printf("Roger Bass API démarré sur le port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}
