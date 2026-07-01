package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type URLHandler struct {
	DB *gorm.DB
}

type ShortenRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

func NewURLHandler(db *gorm.DB) *URLHandler {
	return &URLHandler{
		DB: db,
	}
}

func (h *URLHandler) ShortenURL(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := ShortenResponse{
		ShortCode:"abc123",
		ShortURL: "http://localhost:8080/abc123",
	}
	c.JSON(http.StatusOK, response)
}