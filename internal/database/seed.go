package database

import (
	"log"
	"os"

	"github.com/kgermando/porfolio-roger-bass-api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed inserts initial data if tables are empty
func Seed(db *gorm.DB) {
	seedAdmin(db)
	seedWorks(db)
	seedGallery(db)
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.Admin{}).Count(&count)
	if count > 0 {
		return
	}

	// Read initial password from env or use a safe default
	password := os.Getenv("ADMIN_INITIAL_PASSWORD")
	if password == "" {
		password = "rogerbass2026!"
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Erreur bcrypt: %v", err)
		return
	}

	admin := models.Admin{
		Username: "rogerbass",
		Password: string(hashed),
		FullName: "Mukendi Kadiayi Roger Bass",
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("Erreur lors du seeding de l'admin: %v", err)
	} else {
		log.Printf("Admin créé — username: rogerbass, changer le mot de passe après connexion")
	}
}

func seedGallery(db *gorm.DB) {
	var count int64
	db.Model(&models.GalleryPhoto{}).Count(&count)
	if count > 0 {
		return
	}

	photos := []models.GalleryPhoto{
		{
			Src:       "images/rogerbass2.jpeg",
			Alt:       "Roger Bass avec sa guitare acoustique et son casque audio",
			Caption:   "Roger Bass — Guitariste acoustique & basse",
			IsActive:  true,
			SortOrder: 1,
		},
		{
			Src:       "images/rogzrbass1.jpeg",
			Alt:       "Roger Bass tenant sa guitare basse électrique",
			Caption:   `Roger Bass — "Le guitariste de tous les temps"`,
			IsActive:  true,
			SortOrder: 2,
		},
	}

	if err := db.Create(&photos).Error; err != nil {
		log.Printf("Erreur lors du seeding de la galerie: %v", err)
	} else {
		log.Printf("Seeded %d photos de galerie", len(photos))
	}
}

func seedWorks(_ *gorm.DB) {
	// Les œuvres sont gérées via l'interface admin — aucun seed automatique.
	// Connectez-vous à /admin/login pour ajouter vos vidéos YouTube.
}
