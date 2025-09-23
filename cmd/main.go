package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rifqi535/expense-tracker-api/internal/config"
	"github.com/rifqi535/expense-tracker-api/internal/handlers"
	"github.com/rifqi535/expense-tracker-api/internal/middleware"
	"github.com/rifqi535/expense-tracker-api/internal/repository"
)

// init() dipanggil otomatis sebelum main()
func init() {
	// coba load file .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found, fallback ke system environment")
	}
}

func main() {
	// 🔹 Ambil konfigurasi dari .env
	port := getEnv("PORT", "8081")
	dsn := getEnv("DB_DSN", "")
	if dsn == "" {
		log.Fatal("❌ DB_DSN tidak ditemukan di environment")
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		log.Fatal("❌ JWT_SECRET tidak ditemukan di environment")
	}
	log.Println("✅ JWT_SECRET terbaca")

	// ⬇️ Tambahin ini
	cfg := config.Load()
	middleware.InitJWT(cfg.JWTSecret)
	// ⬆️

	// 🔹 Koneksi ke database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ gagal connect DB: %v", err)
	}
	fmt.Println("✅ Database connected")

	// 🔹 Setup Gin & route
	r := gin.Default()

	// repo & handler
	categoryRepo := repository.NewCategoryRepo(db)
	expenseRepo := repository.NewExpenseRepo(db)
	authHandler := handlers.NewAuthHandler(db)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	expHandler := handlers.NewExpenseHandler(expenseRepo)

	// 🔹 auth routes (public)
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	// 🔹 user routes (protected)
	userRoutes := r.Group("/user")
	userRoutes.Use(middleware.AuthMiddleware())
	{
		userRoutes.GET("/profile", authHandler.GetProfile)
	}

	// 🔹 protected routes
	api := r.Group("/")
	api.Use(middleware.AuthMiddleware())
	{
		// categories
		api.GET("/categories", categoryHandler.List)
		api.POST("/categories", categoryHandler.Create)
		api.PUT("/categories/:id", categoryHandler.Update)
		api.DELETE("/categories/:id", categoryHandler.Delete)

		// expenses
		api.GET("/expenses", expHandler.List)
		api.POST("/expenses", expHandler.Create)
		api.PUT("/expenses/:id", expHandler.Update)
		api.DELETE("/expenses/:id", expHandler.Delete)
	}

	// Jalankan server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	fmt.Printf("🚀 Server running at http://localhost:%s\n", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ gagal start server: %v", err)
	}
}

// helper ambil env dengan default
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
