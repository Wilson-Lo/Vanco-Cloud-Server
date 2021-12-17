package routers

import (
	"github.com/gin-gonic/gin"
	"net/http"
 //   "os"
   // "fmt"
   //  myMail "app/pkg/app"
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
	/** Login Account **/
    router.POST("/login_account", LoginAccount)
    /** Forgot Password **/
    router.POST("/forgot_password", ForgotPassword)
    /** Reset Password **/
    router.POST("/reset_password", ResetPassword)

   //  mailTo := []string {
    //    "lowilson180@gmail.com",
      //  }
       //郵件主題為"Hello"
     //   subject := "Hello"
       // 郵件正文
      //  body := "Good"
  //  myMail.SendMail(mailTo, subject, body)
   // myMail.SendMailTest()
	return router
}

func index(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", "")
}
