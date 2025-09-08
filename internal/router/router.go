package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rifqi535/expense-tracker-api/internal/handlers"
	"github.com/rifqi535/expense-tracker-api/internal/middleware"
)

func SetupRouter(authHandler *handlers.AuthHandler, catHandler *handlers.CategoryHandler, expHandler *handlers.ExpenseHandler) *gin.Engine {
	r := gin.Default()

	// Auth routes
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	// Protected routes (JWT required)
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())

	auth.GET("/category", catHandler.List)
	auth.POST("/category", catHandler.Create)
	auth.PUT("/category/:id", catHandler.Update)
	auth.DELETE("/category/:id", catHandler.Delete)

	auth.GET("/expenses", expHandler.List)
	auth.POST("/expenses", expHandler.Create)
	auth.PUT("/expenses/:id", expHandler.Update)
	auth.DELETE("/expenses/:id", expHandler.Delete)

	return r
}
