package handler

import (
	"net/http"

	"github.com/Scarage1/url-shortener/internal/service"
	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	Service *service.URLService
}

type ShortenRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

func NewURLHandler(service *service.URLService) *URLHandler {
	return &URLHandler{
		Service: service,
	}
}

func (h *URLHandler) RedirectURL(c *gin.Context) {

	shortCode := c.Param("code")

	url, err := h.Service.GetOriginalURL(shortCode)

	if err != nil {

		c.JSON(
			http.StatusNotFound,
			gin.H{
				"error": "URL not found",
			},
		)

		return
	}

	c.Redirect(
		http.StatusFound,
		url.OriginalURL,
	)
}

func (h *URLHandler) ShortenURL(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := h.Service.CreateShortURL(req.URL)

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := ShortenResponse{
		ShortCode: url.ShortCode,
		ShortURL:  "http://localhost:8080/" + url.ShortCode,
	}
	c.JSON(http.StatusOK, response)
}

func (h *URLHandler) GetStats(
	c *gin.Context,
) {

	code := c.Param("code")

	url, err :=
		h.Service.GetStats(code)

	if err != nil {

		c.JSON(
			404,
			gin.H{
				"error": "URL not found",
			},
		)

		return
	}

	c.JSON(
		200,
		gin.H{

			"short_code": url.ShortCode,

			"original_url": url.OriginalURL,

			"clicks": url.ClickCount,

			"created_at": url.CreatedAt,

			"last_accessed": url.LastAccessed,
		},
	)
}
