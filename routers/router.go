package routers

import (
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine{
   	router := gin.New()
   	router.Use(gin.Logger())
   	router.Use(gin.Recovery())
	router.POST("/balance", Connect)
	router.GET("/ws", WSSConnect)
	return router
}