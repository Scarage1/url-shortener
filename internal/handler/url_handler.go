package handler

import (
	"net/http"

	"github.com/Scarage1/url-shortener/internal/service"
	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	Service *service.URLService
	BaseURL string
}

type ShortenRequest struct {
	URL string `json:"url" binding:"required,url"`
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

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid URL",
			},
		)

		return
	}

	userIDValue, exists := c.Get("user_id")

	if !exists {

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "unauthorized",
			},
		)

		return
	}

	userID, ok := userIDValue.(uint)

	if !ok {

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "invalid user",
			},
		)

		return
	}

	url, err :=
		h.Service.CreateShortURL(
			req.URL,
			userID,
		)

	if err != nil {

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	response :=
		ShortenResponse{

			ShortCode: url.ShortCode,

			ShortURL: h.BaseURL + "/" +
				url.ShortCode,
		}

	c.JSON(
		http.StatusOK,
		response,
	)
}

func (h *URLHandler) GetStats(
	c *gin.Context,
) {

	code :=
		c.Param(
			"code",
		)

	userIDValue, exists := c.Get("user_id")

	if !exists {

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "unauthorized",
			},
		)

		return
	}

	userID, ok := userIDValue.(uint)

	if !ok {

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "invalid user",
			},
		)

		return
	}

	url, err :=
		h.Service.GetStats(
			code,
			userID,
		)

	if err != nil {

		c.JSON(
			http.StatusNotFound,
			gin.H{
				"error": "URL not found",
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{

			"short_code": url.ShortCode,

			"original_url": url.OriginalURL,

			"clicks": url.ClickCount,

			"created_at": url.CreatedAt,

			"last_accessed": url.LastAccessed,
		},
	)
}
