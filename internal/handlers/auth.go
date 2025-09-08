package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rifqi535/expense-tracker-api/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("JWT_SECRET")

type AuthHandler struct {
	DB *pgxpool.Pool
}

func NewAuthHandler(db *pgxpool.Pool) *AuthHandler {
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

	_, err := h.DB.Exec(context.Background(),
		`INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		userID, req.Name, req.Email, hashed, time.Now(), time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID uuid.UUID
	var hashedPassword string
	err := h.DB.QueryRow(context.Background(),
		"SELECT id, password_hash FROM users WHERE email=$1", req.Email,
	).Scan(&userID, &hashedPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email/password"})
		return
	}

	if !checkPasswordHash(req.Password, hashedPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email/password"})
		return
	}

	token, err := middleware.GenerateToken(userID.String())
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
	var name, email string
	err = h.DB.QueryRow(context.Background(),
		"SELECT name, email FROM users WHERE id=$1", uid,
	).Scan(&name, &email)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// balikin profile
	c.JSON(http.StatusOK, gin.H{
		"id":    uid.String(),
		"name":  name,
		"email": email,
	})
}
