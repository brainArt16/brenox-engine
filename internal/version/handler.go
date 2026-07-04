package version

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Get(c *gin.Context) {
	c.JSON(http.StatusOK, Get())
}
