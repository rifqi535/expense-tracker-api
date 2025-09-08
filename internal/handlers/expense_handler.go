package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rifqi535/expense-tracker-api/internal/models"
	"github.com/rifqi535/expense-tracker-api/internal/repository"
)

type ExpenseHandler struct {
	Repo *repository.ExpenseRepo
}

func NewExpenseHandler(repo *repository.ExpenseRepo) *ExpenseHandler {
	return &ExpenseHandler{Repo: repo}
}

func (h *ExpenseHandler) List(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	uid, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	// --- PAGINATION ---
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// --- SORTING ---
	sortBy := c.DefaultQuery("sort_by", "date") // "date" atau "amount"
	order := c.DefaultQuery("order", "desc")    // "asc" atau "desc"
	if sortBy != "date" && sortBy != "amount" {
		sortBy = "date"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// --- FILTERS ---
	var (
		categoryID *uuid.UUID
		startDate  *time.Time
		endDate    *time.Time
	)

	if cid := c.Query("category_id"); cid != "" {
		if parsed, err := uuid.Parse(cid); err == nil {
			categoryID = &parsed
		}
	}
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = &t
		}
	}
	if e := c.Query("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			endDate = &t
		}
	}

	// --- QUERY KE REPO ---
	expenses, err := h.Repo.List(c, uid, categoryID, startDate, endDate, limit, offset, sortBy, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"limit":    limit,
		"expenses": expenses,
	})
}

// Create new expense
func (h *ExpenseHandler) Create(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	uid, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Title       string  `json:"title"`
		Amount      float64 `json:"amount"`
		CategoryID  string  `json:"category_id"`
		Description string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	exp := &models.Expense{
		ID:          uuid.New(),
		Title:       req.Title,
		Amount:      req.Amount,
		CategoryID:  categoryID,
		UserID:      uid,
		Description: &req.Description,
	}
	if err := h.Repo.Create(c, exp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, exp)
}

// Update expense
func (h *ExpenseHandler) Update(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	uid, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense id"})
		return
	}

	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Amount      float64 `json:"amount"`
		CategoryID  string  `json:"category_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	okRepo, err := h.Repo.Update(
		c,
		uid,
		id,
		req.Title,
		req.Description,
		req.Amount,
		categoryID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !okRepo {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Expense updated"})
}

// Delete expense
func (h *ExpenseHandler) Delete(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	uid, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense id"})
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

	c.JSON(http.StatusOK, gin.H{"message": "Expense deleted"})
}
