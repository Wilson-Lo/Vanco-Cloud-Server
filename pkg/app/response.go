package app

import (
	"app/models"
	"github.com/gin-gonic/gin"
)

type Gin struct {
	C *gin.Context
}

// Response setting gin.JSON
func (g *Gin) Response(httpCode int, command models.Command) {
	g.C.JSON(httpCode, command)
	return
}
