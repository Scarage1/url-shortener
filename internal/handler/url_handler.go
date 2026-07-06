package handler

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Scarage1/url-shortener/internal/service"
	"github.com/Scarage1/url-shortener/internal/utils"
	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	Service *service.URLService
	BaseURL string
}

type ShortenRequest struct {
	URL   string               `json:"url" binding:"required,url"`
	Rules []ShortenRuleRequest `json:"rules"`
}

type ShortenRuleRequest struct {
	Type       string            `json:"type"`
	Password   string            `json:"password"`
	ActiveFrom string            `json:"active_from"` // ISO 8601 e.g. "2026-07-10T00:00:00Z"
	ExpiresAt  string            `json:"expires_at"`  // ISO 8601
	GeoRoutes  map[string]string `json:"geo_routes"`  // country_code → URL
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

func NewURLHandler(service *service.URLService, baseURL string) *URLHandler {
	return &URLHandler{
		Service: service,
		BaseURL: baseURL,
	}
}

func (h *URLHandler) RedirectURL(c *gin.Context) {

	shortCode := c.Param("code")
	password := c.GetHeader("X-Link-Password")

	if password == "" {
		password = c.Query("password")
	}

	url, err := h.Service.GetOriginalURL(shortCode, password, c.ClientIP())

	if err != nil {

		switch {
		case errors.Is(err, service.ErrPasswordRequired):
			c.JSON(
				http.StatusUnauthorized,
				gin.H{"error": "password required"},
			)
			return
		case errors.Is(err, service.ErrInvalidPassword):
			c.JSON(
				http.StatusUnauthorized,
				gin.H{"error": "invalid password"},
			)
			return
		case errors.Is(err, service.ErrNotYetActive):
			c.JSON(
				http.StatusForbidden,
				gin.H{"error": "link not yet active"},
			)
			return
		case errors.Is(err, service.ErrExpired):
			c.JSON(
				http.StatusGone,
				gin.H{"error": "link has expired"},
			)
			return
		}

		c.JSON(
			http.StatusNotFound,
			gin.H{"error": "URL not found"},
		)
		return
	}

	c.Redirect(http.StatusFound, url.OriginalURL)
}

func (h *URLHandler) ShortenURL(c *gin.Context) {

	var req ShortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid URL"},
		)
		return
	}

	userID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "unauthorized"},
		)
		return
	}

	url, err := h.Service.CreateShortURL(
		req.URL,
		userID,
		toCreateRuleInputs(req.Rules),
	)

	if err != nil {
		if errors.Is(err, service.ErrUnsafeURL) {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "unsafe URL"},
			)
			return
		}
		if errors.Is(err, service.ErrInvalidRule) {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid rules"},
			)
			return
		}
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		ShortenResponse{
			ShortCode: url.ShortCode,
			ShortURL:  h.BaseURL + "/" + url.ShortCode,
		},
	)
}

func toCreateRuleInputs(rules []ShortenRuleRequest) []service.CreateRuleInput {

	if len(rules) == 0 {
		return nil
	}

	inputs := make([]service.CreateRuleInput, 0, len(rules))

	for _, rule := range rules {

		input := service.CreateRuleInput{
			Type:      rule.Type,
			Password:  rule.Password,
			GeoRoutes: rule.GeoRoutes,
		}

		if rule.ActiveFrom != "" {
			if t, err := time.Parse(time.RFC3339, rule.ActiveFrom); err == nil {
				input.ActiveFrom = &t
			}
		}

		if rule.ExpiresAt != "" {
			if t, err := time.Parse(time.RFC3339, rule.ExpiresAt); err == nil {
				input.ExpiresAt = &t
			}
		}

		inputs = append(inputs, input)
	}

	return inputs
}

func (h *URLHandler) GetStats(c *gin.Context) {

	code := c.Param("code")

	userID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "unauthorized"},
		)
		return
	}

	url, err := h.Service.GetStats(code, userID)
	if err != nil {
		c.JSON(
			http.StatusNotFound,
			gin.H{"error": "URL not found"},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"short_code":    url.ShortCode,
			"original_url":  url.OriginalURL,
			"clicks":        url.ClickCount,
			"created_at":    url.CreatedAt,
			"last_accessed": url.LastAccessed,
		},
	)
}

func (h *URLHandler) GetUserLinks(c *gin.Context) {

	userID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "unauthorized"},
		)
		return
	}

	urls, err := h.Service.GetUserLinks(userID)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		return
	}

	c.JSON(http.StatusOK, urls)
}

func (h *URLHandler) DeleteURL(c *gin.Context) {

	code := c.Param("code")

	userID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "unauthorized"},
		)
		return
	}

	err = h.Service.DeleteURL(code, userID)
	if err != nil {
		c.JSON(
			http.StatusNotFound,
			gin.H{"error": "URL not found"},
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ---------------------------------------------------------------------------
// Phase 23E — CSV Import / Export
// ---------------------------------------------------------------------------

// ExportLinks streams the authenticated user's links as a CSV file.
//
//	GET /api/v1/export
func (h *URLHandler) ExportLinks(c *gin.Context) {

	userID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "unauthorized"},
		)
		return
	}

	urls, err := h.Service.GetUserLinks(userID)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="links.csv"`)

	w := csv.NewWriter(c.Writer)

	// Header row
	_ = w.Write([]string{
		"short_code",
		"original_url",
		"clicks",
		"created_at",
	})

	for _, url := range urls {
		_ = w.Write([]string{
			url.ShortCode,
			url.OriginalURL,
			strconv.Itoa(url.ClickCount),
			url.CreatedAt.Format(time.RFC3339),
		})
	}

	w.Flush()
}

// ImportLinks reads a multipart CSV file and batch-creates short URLs.
//
//	POST /api/v1/import   (form-data key: "file")
//
// Each non-header row must have at least one column: the original URL.
// The second column (short_code) is ignored — new codes are generated.
// Returns a JSON summary: { created, skipped, failed }.
func (h *URLHandler) ImportLinks(c *gin.Context) {

	userID, err := utils.GetUserID(c)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "unauthorized"},
		)
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "file is required (multipart key: file)"},
		)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // allow variable columns

	// Skip header row
	if _, err := reader.Read(); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "CSV must have at least a header row"},
		)
		return
	}

	var created, skipped, failed int

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			failed++
			continue
		}
		if len(record) == 0 || record[0] == "" {
			skipped++
			continue
		}

		originalURL := record[0]

		_, err = h.Service.CreateShortURL(originalURL, userID, nil)
		if err != nil {
			if errors.Is(err, service.ErrUnsafeURL) {
				failed++
			} else {
				// Duplicate URL → returns the existing code (skipped, not an error)
				skipped++
			}
			continue
		}

		created++
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"created": created,
			"skipped": skipped,
			"failed":  failed,
		},
	)
}
