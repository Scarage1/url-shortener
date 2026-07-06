package handler

import (
	"net/http"

	"github.com/Scarage1/url-shortener/internal/service"
	"github.com/Scarage1/url-shortener/internal/utils"
	"github.com/gin-gonic/gin"
)

// DashboardHandler serves the aggregated dashboard endpoint.
type DashboardHandler struct {
	Service *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{Service: svc}
}

// GetDashboard returns aggregated stats for the authenticated user's org.
//
//	GET /api/v1/dashboard
func (h *DashboardHandler) GetDashboard(c *gin.Context) {

	orgID, err := utils.GetOrgID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.Service.GetDashboard(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load dashboard"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
