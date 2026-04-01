package handlers

import (
	"strconv"

	"github.com/chengtb/restaurant-kds/internal/middleware"
	"github.com/chengtb/restaurant-kds/internal/services"
	"github.com/gin-gonic/gin"
)

type GiftHandler struct {
	GiftService *services.GiftService
}

func NewGiftHandler(gs *services.GiftService) *GiftHandler {
	return &GiftHandler{GiftService: gs}
}

// CreateGift handles gift requests
func (h *GiftHandler) CreateGift(c *gin.Context) {
	var req services.GiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	record, err := h.GiftService.CreateGift(&req)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.SuccessWithMessage(c, "Gift created", record)
}

// ListGifts lists gift records
func (h *GiftHandler) ListGifts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := h.GiftService.ListGifts(page, pageSize)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"records":   records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
