package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rifqi535/expense-tracker-api/internal/models"
	"github.com/rifqi535/expense-tracker-api/internal/repository"
)

type CategoryHandler struct {
	Repo *repository.CategoryRepo
}

func NewCategoryHandler(repo *repository.CategoryRepo) *CategoryHandler {
	return &CategoryHandler{Repo: repo}
}

// List categories for logged-in user
func (h *CategoryHandler) List(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	uid, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	categories, err := h.Repo.ListByUser(c, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// Create new category
func (h *CategoryHandler) Create(c *gin.Context) {
	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("user_id")
	uid, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	category := &models.Category{
		ID:     uuid.New(),
		Title:  req.Title,
		UserID: uid,
	}
	if err := h.Repo.Create(c, category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, category)
}

// Update category
func (h *CategoryHandler) Update(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	uid, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	okRepo, err := h.Repo.Update(c, uid, id, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !okRepo {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category updated"})
}

// Delete category
func (h *CategoryHandler) Delete(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	uid, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	okRepo, err := h.Repo.Delete(c, uid, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !okRepo {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}
