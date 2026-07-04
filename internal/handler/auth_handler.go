package handler

import (
	"net/http"

	"github.com/Scarage1/url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Service *service.AuthService
}

func NewAuthHandler(
	service *service.AuthService,
) *AuthHandler {

	return &AuthHandler{
		Service: service,
	}
}

type AuthRequest struct {
	Email string `json:"email" binding:"required,email"`

	Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthHandler) Register(
	c *gin.Context,
) {

	var req AuthRequest

	if err := c.ShouldBindJSON(
		&req,
	); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid request",
			},
		)

		return
	}

	err :=
		h.Service.Register(
			req.Email,
			req.Password,
		)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	c.JSON(
		http.StatusCreated,
		gin.H{
			"message": "user created",
		},
	)
}

func (h *AuthHandler) Login(
	c *gin.Context,
) {

	var req AuthRequest

	if err := c.ShouldBindJSON(
		&req,
	); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid request",
			},
		)

		return
	}

	token, err :=
		h.Service.Login(
			req.Email,
			req.Password,
		)

	if err != nil {

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"token": token,
		},
	)
}
