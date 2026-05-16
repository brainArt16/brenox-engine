package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
	Handler layer:
	- receives HTTP request
	- validates/parses request
	- calls service layer
	- returns JSON response
*/

type Handler struct {
	service *Service
}

func NewHandler(
	service *Service,
) *Handler {

	return &Handler{
		service: service,
	}
}

// Register endpoint handler.
func (h *Handler) Register(
	c *gin.Context,
) {

	var req RegisterRequest

	// Bind incoming JSON body into Go struct.
	err := c.ShouldBindJSON(&req)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid request body",
			},
		)

		return
	}

	// Call business logic layer.

	user, err := h.service.Register(
		c.Request.Context(),
		req,
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

	// Return successful response.

	c.JSON(
		http.StatusCreated,
		gin.H{
			"id": user.ID,
			"email": user.Email,
			"username": user.Username,
		},
	)
}


// Login endpoint handler.
func (h *Handler) Login(
	c *gin.Context,
) {

	var req LoginRequest

	err := c.ShouldBindJSON(&req)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid request",
			},
		)

		return
	}

	token, err := h.service.Login(
		c.Request.Context(),
		req,
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