package handlers

import (
	"strconv"

	"github.com/chengtb/restaurant-kds/internal/middleware"
	"github.com/chengtb/restaurant-kds/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	DB *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{DB: db}
}

// ListCategories returns all product categories with products
func (h *ProductHandler) ListCategories(c *gin.Context) {
	var categories []models.Category
	if err := h.DB.Preload("Products", "status = ?", models.ProductStatusActive).
		Order("sort ASC").Find(&categories).Error; err != nil {
		middleware.InternalError(c, "Failed to fetch categories")
		return
	}
	middleware.Success(c, categories)
}

// GetProduct returns a single product with specs
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid product ID")
		return
	}

	var product models.Product
	if err := h.DB.Preload("Specs").Preload("Category").First(&product, id).Error; err != nil {
		middleware.NotFound(c, "Product not found")
		return
	}
	middleware.Success(c, product)
}

// ListProducts returns products with optional category filter
func (h *ProductHandler) ListProducts(c *gin.Context) {
	categoryID := c.Query("category_id")
	query := h.DB.Preload("Specs").Where("status = ?", models.ProductStatusActive)
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	var products []models.Product
	if err := query.Find(&products).Error; err != nil {
		middleware.InternalError(c, "Failed to fetch products")
		return
	}
	middleware.Success(c, products)
}
