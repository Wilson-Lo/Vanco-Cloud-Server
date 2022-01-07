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
	router.POST("/api/pi", Connect)
	router.GET("/api/ws", WSSConnect)
    /** Get User Info **/
    router.POST("/api/user_info", Get_User_Info)
	/** Create Account **/
	router.POST("/api/create_account", CreateAccount)
	/** Login Account **/
    router.POST("/api/login_account", LoginAccount)
    /** Forgot Password **/
    router.POST("/api/forgot_password", ForgotPassword)
    /** Reset Password **/
    router.POST("/api/reset_password", ResetPassword)
    /** All Device List **/
    router.POST("/api/all_device_list", GetAllDeviceList)
    /** Refresh Token **/
    router.POST("/api/refresh_token", Refresh_token)
    /** Logout Account **/
    router.POST("/api/logout_account", Logout_account)
    /** Modify Device Name **/
    router.POST("/api/device_name", Modify_Device_Name)

	return router
}

func index(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", "")
}
