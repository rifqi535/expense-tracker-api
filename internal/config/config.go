package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DB_DSN    string
	JWTSecret string
}

func Load() *Config {
	_ = godotenv.Load()

	c := &Config{
		Port:      getEnv("PORT", "8081"),
		DB_DSN:    getEnv("DB_DSN", "postgres://postgres:uu@localhost:5432/expense_tracker?sslmode=disable"),
		JWTSecret: getEnv("JWT_SECRET", "supersecretultra"),
	}

	if c.JWTSecret == "supersecretultra" {
		log.Println("[WARM] JWT_SECRET menggunakan default, sebaiknya ganti ddi .env")

	}
	return c

}

func getEnv(Key, def string) string {
	if v := os.Getenv(Key); v != "" {
		return v
	}
	return def
}
