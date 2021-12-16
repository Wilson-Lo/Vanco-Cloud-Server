package routers

import (
	e "app/pkg/e"
	b64 "encoding/base64"
    myMail "app/pkg/app"
    myTool "app/pkg/tool"
    passervice "app/service/passervice"
	//jwt "app/pkg/jwt"
	"encoding/json"
	"fmt"
	"app/models"
	"net/http"
    "app/pkg/app"
	"time"
	"strings"
	"github.com/gin-gonic/gin"
	"database/sql"
    "github.com/go-sql-driver/mysql"
)

//mariaDB Config
var cfg = mysql.Config{
        User:   "WilsonLo",
        Passwd:  "Xjij0vu;;",
        Net:    "tcp",
        Addr:   "127.0.0.1:3306",
        DBName: "db_user",
        AllowNativePasswords: true,
}

func Connect(c *gin.Context) {

    appG := app.Gin{C: c}
	var cmd models.Command
	err := c.BindJSON(&cmd)
	if err != nil {
		cmd.Body = "fail"
		appG.Response(http.StatusInternalServerError, cmd)
		return
	}

	if cmd.Method == "cmd" {
		//signKey, err := connectService.GenerateSignKey(cmd)
		//if err != nil {
		//	cmd.Body = e.FAILURE
		passervice.AddToGinList("wilson", appG)
        passervice.SendMsgToMachine(cmd.To, string("{\"method\":\"" + cmd.Method + "\"}"))

        ch := make(chan bool)
        passervice.AddToChanMap("wilson", ch)
        select {
           case <-ch:
        		break
           case <-time.After(10 * time.Second):
        		cmdRes := models.Command{}
        		cmdRes.Etag = cmd.Etag
        		cmdRes.Body = "Timeout"
        		appG.Response(http.StatusRequestTimeout, cmdRes)
        		break
        }
	}
		//else {
		//cmd.Body = "1.0.0"
		//passervice.AddToGinList("wilson", appG)
		//passervice.SendMsgToMachine(cmd.To, string("{\"method\":\"" + cmd.Method + "\"}"))
			//cmd.Extra = signKey
		//}
	//	appG.Response(http.StatusOK, cmd)


	//}
}

//create new account
func CreateAccount(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
	err := c.BindJSON(&cmd)

	if err != nil {
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
        cmd.Sign = getSign(cmd)
        appG.Response(http.StatusOK, cmd)
		return
	}


	 var sign = getSign(cmd)

     //check sign value
     if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
     }else{
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
        cmd.Sign = getSign(cmd)
        appG.Response(http.StatusOK, cmd)
        return
     }

     var bodyData = strings.ReplaceAll(cmd.Body, e.SaltFirst, "")
     bodyData = strings.ReplaceAll(bodyData, e.SaltAfter, "")
     bytes, err := b64.StdEncoding.DecodeString(bodyData)

    //get new account info
	var accountInfo = models.CmdCreateAccount{}
    json.Unmarshal(bytes, &accountInfo)

    //Valid E-mail
    if(!myTool.ValidEmail(accountInfo.Account)){
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Email address format not correct !\"}")
       cmd.Sign = getSign(cmd)
       appG.Response(http.StatusOK, cmd)
       return
    }

    // Create the database handle, confirm driver is present
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if(err != nil){
	   fmt.Println("Connect to DB Failed !")
	   defer db.Close()
	}

    // See "Important settings" section.
    db.SetConnMaxLifetime(time.Minute * 3)
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(10)

    //check email is not register before
   // fmt.Println("SELECT * FROM users WHERE account = '" + accountInfo.Account + "'")
	sql := fmt.Sprintf("SELECT * FROM users WHERE account = '" + accountInfo.Account + "'")
    rows, err := db.Query(sql)
    if err != nil {
       fmt.Println("SQLite occur error : " + err.Error())
       return
    }
    defer db.Close()

    var isEmailUsed = false
    for rows.Next() {
        isEmailUsed = true
    }

    if(!isEmailUsed){
       //add new account to DB
       var pwMD5 = myTool.ToMD5(accountInfo.Password)
       cmd.Body = "successful"
       dt := time.Now()
       formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
               dt.Year(), dt.Month(), dt.Day(),
               dt.Hour(), dt.Minute(), dt.Second())
     //  fmt.Println("INSERT INTO users (account, password, role, time) VALUES (\"" + accountInfo.Account + "\", \"" + accountInfo.Password + "\", 2, \""+ formatted +"\")")
       _, err := db.Exec("INSERT INTO users (account, password, role, time) VALUES (\"" + accountInfo.Account + "\", \"" + pwMD5 + "\", 2, \""+ formatted +"\")")
       if err != nil {
       		cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Add new account failed\"}")
            cmd.Sign = getSign(cmd)
            appG.Response(http.StatusOK, cmd)
       		return
       }

       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Register Successful !\"}")
       cmd.Sign = getSign(cmd)
       appG.Response(http.StatusOK, cmd)
    }else{
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" This E-Mail has been registered\"}")
       cmd.Sign = getSign(cmd)
       appG.Response(http.StatusOK, cmd)
    }
}


//login account
func LoginAccount(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
	err := c.BindJSON(&cmd)

	if err != nil {
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
       cmd.Sign = getSign(cmd)
       appG.Response(http.StatusOK, cmd)
		return
	}

    var sign = getSign(cmd)

    if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
    }else{
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
       cmd.Sign = getSign(cmd)
       appG.Response(http.StatusOK, cmd)
       return
    }

    var bodyData = strings.ReplaceAll(cmd.Body, e.SaltFirst, "")
    bodyData = strings.ReplaceAll(bodyData, e.SaltAfter, "")
    bytes, err := b64.StdEncoding.DecodeString(bodyData)

	//get login info
    var loginInfo = models.CmdCreateAccount{}
    json.Unmarshal(bytes, &loginInfo)
    //fmt.Println("loginInfo = " + string(loginInfo))
    fmt.Println("account = " + loginInfo.Account + " - password = " + loginInfo.Password )
    // Create the database handle, confirm driver is present
    db, err := sql.Open("mysql", cfg.FormatDSN())
    if(err != nil){
        fmt.Println("Connect to DB Failed !")
    }
    defer db.Close()

	sql := fmt.Sprintf("SELECT * FROM users WHERE account = '" + loginInfo.Account + "';")
    rows, err := db.Query(sql)
    if err != nil {
       fmt.Println("SQLite occur error : " + err.Error())
       return
    }
    defer rows.Close()

    var  id int
    var  accounts string
    var  password string
    var  role int
    var  time string

    if(rows.Next()) {
        err := rows.Scan(&id, &accounts, &password, &role, &time)
        if err != nil {
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Account or Password not correct !\"}")
           cmd.Sign = getSign(cmd)
           appG.Response(http.StatusInternalServerError, cmd)
           return
        }

       //check password
       var userEnterPassword = myTool.ToMD5(loginInfo.Password)
       if(strings.Compare(userEnterPassword, password) == 0){
         // token, err := jwt.CreateToken(uint64(id))
        //  if(err != nil){
         //   fmt.Println("create token error : ", err.Error())
         // }
       //   fmt.Println("token : ", token)
          cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Login Successful !\" , \"token\": \"" + "token" + "\"}")
          cmd.Sign = getSign(cmd)
          appG.Response(http.StatusOK, cmd)
       }else{
          cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Account or Password not correct !\"}")
          cmd.Sign = getSign(cmd)
          appG.Response(http.StatusOK, cmd)
       }
    }else{
       //account not exist
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Account or Password not correct !\"}")
       cmd.Sign = getSign(cmd)
       appG.Response(http.StatusOK, cmd)
    }
}

/**
*  Get sign value
*/
func getSign(data models.Command) string{
     var allData = "body="+data.Body+"&etag="+data.Etag+"&extra="+data.Extra+"&method="+data.Method+"&time="+data.Time+"&to="+data.To
     return myTool.ToMD5(allData)
}

//forgot password
func ForgotPassword(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
	err := c.BindJSON(&cmd)

	if err != nil {
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
       cmd.Sign = getSign(cmd)
       appG.Response(http.StatusOK, cmd)
		return
	}

    var sign = getSign(cmd)

    if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
    }else{
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
       cmd.Sign = getSign(cmd)
       appG.Response(http.StatusOK, cmd)
       return
    }

    var bodyData = strings.ReplaceAll(cmd.Body, e.SaltFirst, "")
    bodyData = strings.ReplaceAll(bodyData, e.SaltAfter, "")
    bytes, err := b64.StdEncoding.DecodeString(bodyData)

	//get account
    var accountInfo = models.CmdForgotPassowrd{}
    json.Unmarshal(bytes, &accountInfo)
    //fmt.Println("loginInfo = " + string(loginInfo))
    fmt.Println("account = " + accountInfo.Account)
    // Create the database handle, confirm driver is present
    db, err := sql.Open("mysql", cfg.FormatDSN())
    if(err != nil){
        fmt.Println("Connect to DB Failed !")
    }
    defer db.Close()

	sql := fmt.Sprintf("SELECT * FROM users WHERE account = '" + accountInfo.Account + "';")
    rows, err := db.Query(sql)
    if err != nil {
       fmt.Println("SQLite occur error : " + err.Error())
       return
    }
    defer rows.Close()

    var  id int
    var  accounts string
    var  password string
    var  role int
    var  time string

    if(rows.Next()) {
        err := rows.Scan(&id, &accounts, &password, &role, &time)
        if err != nil {
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
           cmd.Sign = getSign(cmd)
           appG.Response(http.StatusInternalServerError, cmd)
           return
        }
        fmt.Println("prepare to create token")
        myMail.SendMail(accountInfo.Account, "Hi Test Reset password mail link !")
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"The reset password link is sent to your mail !\"}")
        cmd.Sign = getSign(cmd)
        appG.Response(http.StatusOK, cmd)

    }else{
       //account not exist
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Account is not exist !\"}")
       cmd.Sign = getSign(cmd)
       appG.Response(http.StatusOK, cmd)
    }
}