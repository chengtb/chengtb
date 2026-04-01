package handlers

import (
	"strconv"

	"github.com/chengtb/restaurant-kds/internal/middleware"
	"github.com/chengtb/restaurant-kds/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TableHandler struct {
	DB *gorm.DB
}

func NewTableHandler(db *gorm.DB) *TableHandler {
	return &TableHandler{DB: db}
}

// ListRegions returns all dining regions with tables
func (h *TableHandler) ListRegions(c *gin.Context) {
	var regions []models.Region
	if err := h.DB.Preload("Tables").Find(&regions).Error; err != nil {
		middleware.InternalError(c, "Failed to fetch regions")
		return
	}
	middleware.Success(c, regions)
}

// GetTable returns table info
func (h *TableHandler) GetTable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid table ID")
		return
	}

	var table models.Table
	if err := h.DB.Preload("Region").First(&table, id).Error; err != nil {
		middleware.NotFound(c, "Table not found")
		return
	}
	middleware.Success(c, table)
}
