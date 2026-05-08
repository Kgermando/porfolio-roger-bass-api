package handlers

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kgermando/porfolio-roger-bass-api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	db *gorm.DB
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input models.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Corps de requête invalide"})
	}

	if input.Username == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Identifiants requis"})
	}

	var admin models.Admin
	if err := h.db.Where("username = ?", input.Username).First(&admin).Error; err != nil {
		// Use a constant-time failure message to avoid enumeration attacks
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Identifiants incorrects"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(input.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Identifiants incorrects"})
	}

	token, err := generateToken(admin.ID, admin.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur serveur"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"admin": fiber.Map{
			"id":        admin.ID,
			"username":  admin.Username,
			"full_name": admin.FullName,
		},
	})
}

// Me handles GET /api/auth/me — returns current authenticated admin
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	adminID := c.Locals("adminID")
	var admin models.Admin
	if err := h.db.First(&admin, adminID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Administrateur introuvable"})
	}
	return c.JSON(fiber.Map{
		"id":        admin.ID,
		"username":  admin.Username,
		"full_name": admin.FullName,
	})
}

func generateToken(adminID uint, username string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"sub":      adminID,
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
