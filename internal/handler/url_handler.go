package handler

import (
	"errors"
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
	URL   string               `json:"url" binding:"required,url"`
	Rules []ShortenRuleRequest `json:"rules"`
}

type ShortenRuleRequest struct {
	Type     string `json:"type"`
	Password string `json:"password"`
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

	url, err := h.Service.GetOriginalURL(shortCode, password)

	if err != nil {

		switch {
		case errors.Is(err, service.ErrPasswordRequired):
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "password required",
				},
			)
			return
		case errors.Is(err, service.ErrInvalidPassword):
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid password",
				},
			)
			return
		}

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
			toCreateRuleInputs(req.Rules),
		)

	if err != nil {

		if errors.Is(err, service.ErrUnsafeURL) {

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": "unsafe URL",
				},
			)

			return
		}

		if errors.Is(err, service.ErrInvalidRule) {

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": "invalid rules",
				},
			)

			return
		}

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

func toCreateRuleInputs(
	rules []ShortenRuleRequest,
) []service.CreateRuleInput {

	if len(rules) == 0 {
		return nil
	}

	inputs := make([]service.CreateRuleInput, 0, len(rules))

	for _, rule := range rules {
		inputs = append(
			inputs,
			service.CreateRuleInput{
				Type:     rule.Type,
				Password: rule.Password,
			},
		)
	}

	return inputs
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
