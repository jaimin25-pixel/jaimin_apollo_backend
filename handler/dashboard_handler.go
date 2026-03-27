package handler

import (
	"net/http"

	"apollo-backend/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashSvc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashSvc: svc}
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID, _ := c.Get("userID")
	data, err := h.dashSvc.GetDashboard(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard"})
		return
	}
	c.JSON(http.StatusOK, data)
}
