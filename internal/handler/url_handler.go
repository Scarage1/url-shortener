package handler

import (
	"net/http"

	"github.com/Scarage1/url-shortener/internal/service"
	"github.com/Scarage1/url-shortener/internal/utils"
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

	userID, err := utils.GetUserID(c)

	if err != nil {

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "unauthorized",
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

	code := c.Param("code")

	userID, err := utils.GetUserID(c)

	if err != nil {

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "unauthorized",
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

func (h *URLHandler) GetUserLinks(
	c *gin.Context,
) {

	userID, err := utils.GetUserID(c)

	if err != nil {

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "unauthorized",
			},
		)

		return
	}

	urls, err :=
		h.Service.GetUserLinks(
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

	c.JSON(
		http.StatusOK,
		urls,
	)
}

func (h *URLHandler) DeleteURL(
	c *gin.Context,
) {

	code := c.Param("code")

	userID, err := utils.GetUserID(c)

	if err != nil {

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "unauthorized",
			},
		)

		return
	}

	err = h.Service.DeleteURL(
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
			"message": "deleted",
		},
	)
}
