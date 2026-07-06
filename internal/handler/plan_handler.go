package handler

import (
	"net/http"

	"github.com/Scarage1/url-shortener/internal/service"
	"github.com/Scarage1/url-shortener/internal/utils"
	"github.com/gin-gonic/gin"
)

// PlanHandler exposes plan and usage information for the dashboard.
type PlanHandler struct {
	QuotaService *service.QuotaService
}

func NewPlanHandler(quotaService *service.QuotaService) *PlanHandler {
	return &PlanHandler{QuotaService: quotaService}
}

// GetPlan returns the current plan and its limits for the authenticated user's org.
//
//	GET /api/v1/plan
func (h *PlanHandler) GetPlan(c *gin.Context) {

	orgID, err := utils.GetOrgID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	plan, err := h.QuotaService.GetPlanForOrg(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":               plan.Name,
		"display_name":       plan.DisplayName,
		"max_links":          plan.MaxLinks,
		"max_redirects":      plan.MaxRedirects,
		"max_api_calls":      plan.MaxAPICalls,
		"max_domains":        plan.MaxDomains,
		"max_geo_rules":      plan.MaxGeoRules,
		"max_password_links": plan.MaxPasswordLinks,
		"max_schedule_links": plan.MaxScheduleLinks,
		"max_members":        plan.MaxMembers,
		"rate_limit":         plan.RateLimit,
		"price_monthly":      plan.PriceMonthly,
		"price_yearly":       plan.PriceYearly,
	})
}

// GetUsage returns current month usage stats for the authenticated user's org.
//
//	GET /api/v1/usage
func (h *PlanHandler) GetUsage(c *gin.Context) {

	orgID, err := utils.GetOrgID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	usage, err := h.QuotaService.GetUsage(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load usage"})
		return
	}

	c.JSON(http.StatusOK, usage)
}
