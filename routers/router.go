package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/gzip"
	"net/http"
 //   "os"
   // "fmt"
   //  myMail "app/pkg/app"
)

func InitRouter() *gin.Engine{

   	router := gin.New()
   	router.Use(gin.Logger())
   	router.Use(gin.Recovery())
   	router.Use(gzip.Gzip(gzip.DefaultCompression))
    router.LoadHTMLFiles("asset/index.html")
   	router.Static("/asset", "asset")
   	router.GET("/", index)
	router.POST("/pi", Connect)
	router.GET("/ws", WSSConnect)

	/** Create Account **/
	router.POST("/create_account", CreateAccount)
	/** Login Account **/
    router.POST("/login_account", LoginAccount)
    /** Forgot Password **/
    router.POST("/forgot_password", ForgotPassword)
    /** Reset Password **/
    router.POST("/reset_password", ResetPassword)
    /** Device List **/
    router.POST("/device_list", GetDeviceList)
    /** Refresh Token **/
    router.POST("/refresh_token", Refresh_token)
    /** Logout Account **/
    router.POST("/logout_account", Logout_account)

	return router
}

func index(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", "")
}
