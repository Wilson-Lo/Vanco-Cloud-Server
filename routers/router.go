package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/gzip"
	"net/http"
    "github.com/gin-contrib/cors"
 //   "os"
    "fmt"
   //  myMail "app/pkg/app"
)

func InitRouter() *gin.Engine{
   	router := gin.New()
    router.Use(cors.New(cors.Config{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{"POST", "PUT", "PATCH", "DELETE"},
        AllowHeaders: []string{"Content-Type,access-control-allow-origin, access-control-allow-headers"},
    }))
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
    /** Get Device list by user **/
    router.POST("/api/device_list", GetDeviceList)
    /** Refresh Token **/
    router.POST("/api/refresh_token", Refresh_token)
    /** Logout Account **/
    router.POST("/api/logout_account", Logout_account)
    /** Modify Device Name **/
    router.POST("/api/device_name", Modify_Device_Name)
    /** Get Associate Code **/
    router.POST("/api/associate_code", Get_Associate_Code)
    /** Remove Device By User **/
    router.POST("/api/remove_device", RemoveDevice)
    /** Add Device Under User Account **/
    router.POST("/api/add_device", AddDevice)
	return router
}

func index(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", "")
}
