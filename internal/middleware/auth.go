// middleware/jwt.go
package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret []byte

// InitJWT harus dipanggil dari main.go setelah .env diload
func InitJWT(secret string) {
	if secret == "" {
		panic("JWT_SECRET is empty, set di .env atau environment")
	}
	jwtSecret = []byte(secret)
	fmt.Println("✅ JWT_SECRET loaded")
}

// GenerateToken bikin JWT dari userID
func GenerateToken(userID string) (string, error) {
	if len(jwtSecret) == 0 {
		return "", errors.New("JWT_SECRET is not set")
	}

	claims := jwt.MapClaims{
		"user_id": fmt.Sprintf("%s", userID),
		"exp":     time.Now().Add(72 * time.Hour).Unix(), // 3 hari
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken untuk validasi & extract user_id dari token
func ParseToken(tokenString string) (uuid.UUID, error) {
	if len(jwtSecret) == 0 {
		return uuid.Nil, errors.New("JWT_SECRET is not set")
	}

	// ✅ parse & validasi
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// cek signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		// 🔎 kasih detail biar ketahuan error aslinya
		return uuid.Nil, fmt.Errorf("token parse error: %w", err)
	}
	if !token.Valid {
		return uuid.Nil, errors.New("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user id, got type %T, value: %#v", claims["user_id"], claims["user_id"])
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	return uid, nil
}

// AuthMiddleware untuk validasi JWT di header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := ParseToken(tokenStr)
		if err != nil {
			// 🔎 log biar keliatan di console error sebenarnya
			fmt.Println("❌ JWT error:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// simpan user_id ke context
		c.Set("user_id", userID)
		c.Next()
	}
}
