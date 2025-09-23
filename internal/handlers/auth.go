package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rifqi535/expense-tracker-api/internal/middleware"
	"github.com/rifqi535/expense-tracker-api/internal/models"
	"gorm.io/gorm"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("JWT_SECRET")

type AuthHandler struct {
	DB *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

// 🔑 Hash password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// 🔑 Cek password
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// 🔑 Generate JWT
func GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID, // harus string!
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 📌 Register
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, _ := hashPassword(req.Password)
	userID := uuid.New()

	result := h.DB.WithContext(context.Background()).Exec(
		`INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
	 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, req.Name, req.Email, hashed, time.Now(), time.Now(),
	)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user registered"})

}

// 📌 Login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// ambil user dari DB
	var user models.User
	err := h.DB.WithContext(c.Request.Context()).
		Select("id", "password_hash").
		Where("email = ?", req.Email).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email/password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// cek password
	if !checkPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email/password"})
		return
	}

	// generate token
	token, err := middleware.GenerateToken(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// 📌 GetProfile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// ambil user_id dari context (diset di AuthMiddleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found in context"})
		return
	}

	// paksa ke string apapun tipenya
	userIDStr := fmt.Sprint(userIDVal)

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	// ambil data user dari DB
	var user models.User
	err = h.DB.WithContext(c.Request.Context()).
		Select("name", "email").
		Where("id = ?", uid).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// balikin profile
	c.JSON(http.StatusOK, gin.H{
		"id":    uid.String(),
		"name":  user.Name,
		"email": user.Email,
	})
}
