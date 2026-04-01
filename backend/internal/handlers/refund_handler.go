package handlers

import (
	"strconv"

	"github.com/chengtb/restaurant-kds/internal/middleware"
	"github.com/chengtb/restaurant-kds/internal/services"
	"github.com/gin-gonic/gin"
)

type RefundHandler struct {
	RefundService *services.RefundService
}

func NewRefundHandler(rs *services.RefundService) *RefundHandler {
	return &RefundHandler{RefundService: rs}
}

// CreateRefund handles refund requests
func (h *RefundHandler) CreateRefund(c *gin.Context) {
	var req services.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	record, err := h.RefundService.CreateRefund(&req)
	if err != nil {
		middleware.Forbidden(c, err.Error())
		return
	}

	middleware.SuccessWithMessage(c, "Refund processed", record)
}

// ListRefunds lists refund records
func (h *RefundHandler) ListRefunds(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := h.RefundService.ListRefunds(page, pageSize)
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
