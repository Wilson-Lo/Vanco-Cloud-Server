package routers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func InitRouter() *gin.Engine{
   	router := gin.New()
   	router.Use(gin.Logger())
   	router.Use(gin.Recovery())
    router.LoadHTMLFiles("asset/index.html")
   	router.Static("/asset", "asset")
   	router.GET("/", index)
	router.POST("/pi", Connect)
	router.GET("/ws", WSSConnect)

	/** Create Account **/
	router.POST("/create_account", CreateAccount)
	return router
}

func index(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", "")
}