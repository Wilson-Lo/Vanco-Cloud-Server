package routers

import (
	e "app/pkg/e"
	b64 "encoding/base64"
    myMail "app/pkg/app"
    myTool "app/pkg/tool"
    passervice "app/service/passervice"
    db "app/pkg/db"
	myJwt "app/pkg/jwt"
	"encoding/json"
	"fmt"
	"app/models"
	"net/http"
    "app/pkg/app"
	"time"
	"strings"
	"os"
    "strconv"
	"github.com/gin-gonic/gin"
	"database/sql"
	"github.com/dgrijalva/jwt-go"
)

func Connect(c *gin.Context) {

    appG := app.Gin{C: c}
    var cmd models.Command
	err := c.BindJSON(&cmd)
	if err != nil {
       cmd.Body = "fail"
       appG.Response(http.StatusInternalServerError, cmd)
       return
    }

    //var td *jwt.Todo
    tokenAuth, err := myJwt.ExtractTokenMetadata(c.Request)
    if err != nil {
       fmt.Println("unauthorized 1 ")
       cmd.Body = "{\"result\": \"" + e.FAILURE + "\" , \"message\": \" unauthorized !\"}"
       appG.Response(http.StatusOK, cmd)
       return
     }

    userId, err := myJwt.FetchAuth(tokenAuth)
    if err != nil {
      fmt.Println("unauthorized 2 ")
      cmd.Body = "{\"result\": \"" + e.FAILURE + "\" , \"message\": \" unauthorized !\"}"
      appG.Response(http.StatusOK, cmd)
      return
    }
    fmt.Println("userId " , userId)
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

/**
* Create new account
*/
func CreateAccount(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
	err := c.BindJSON(&cmd)

	if err != nil {
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
        cmd.Sign = myTool.GetSign(cmd)
        appG.Response(http.StatusOK, cmd)
		return
	}


	 var sign = myTool.GetSign(cmd)

     //check sign value
     if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
     }else{
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
        cmd.Sign = myTool.GetSign(cmd)
        appG.Response(http.StatusOK, cmd)
        return
     }

     var bodyData = strings.ReplaceAll(cmd.Body, e.SaltFirst, "")
     bodyData = strings.ReplaceAll(bodyData, e.SaltAfter, "")
     bytes, err := b64.StdEncoding.DecodeString(bodyData)

    //get new account info
	var accountInfo = models.AccountObject{}
    json.Unmarshal(bytes, &accountInfo)

    //Valid E-mail
    if(!myTool.ValidEmail(accountInfo.Account)){
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Email address format not correct !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
       return
    }

    // Create the database handle, confirm driver is present
	db, err := sql.Open("mysql", db.Cfg.FormatDSN())
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
            cmd.Sign = myTool.GetSign(cmd)
            appG.Response(http.StatusOK, cmd)
       		return
       }

       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Register Successful !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
    }else{
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" This E-Mail has been registered\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
    }
}


/**
* Login account
*/
func LoginAccount(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
	err := c.BindJSON(&cmd)

	if err != nil {
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
		return
	}

    var sign = myTool.GetSign(cmd)

    if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
    }else{
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
       return
    }

    var bodyData = strings.ReplaceAll(cmd.Body, e.SaltFirst, "")
    bodyData = strings.ReplaceAll(bodyData, e.SaltAfter, "")
    bytes, err := b64.StdEncoding.DecodeString(bodyData)

	//get login info
    var loginInfo = models.AccountObject{}
    json.Unmarshal(bytes, &loginInfo)
    //fmt.Println("loginInfo = " + string(loginInfo))
    fmt.Println("account = " + loginInfo.Account + " - password = " + loginInfo.Password )
    // Create the database handle, confirm driver is present
    db, err := sql.Open("mysql", db.Cfg.FormatDSN())
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
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusInternalServerError, cmd)
           return
        }

       //check password
       var userEnterPassword = myTool.ToMD5(loginInfo.Password)
       if(strings.Compare(userEnterPassword, password) == 0){
          token, err := myJwt.CreateToken(uint64(id))
          if(err != nil){
            fmt.Println("create token error : ", err.Error())
          }
          saveErr := myJwt.CreateAuth(uint64(id), token)
          if saveErr != nil {
              fmt.Println("save token error : ", saveErr.Error())
          }
          fmt.Println("AccessToken : ", token.AccessToken, " RefreshToken : ", token.RefreshToken)
          cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Login Successful !\" , \"access_token\": \"" + token.AccessToken + "\", \"refresh_token\": \"" + token.RefreshToken + "\"}")
          cmd.Sign = myTool.GetSign(cmd)
          appG.Response(http.StatusOK, cmd)
       }else{
          cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Account or Password not correct !\"}")
          cmd.Sign = myTool.GetSign(cmd)
          appG.Response(http.StatusOK, cmd)
       }
    }else{
       //account not exist
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Account or Password not correct !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
    }
}

/**
* Forgot password
*/
func ForgotPassword(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
	err := c.BindJSON(&cmd)

	if err != nil {
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
		return
	}

    var sign = myTool.GetSign(cmd)

    if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
    }else{
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
       return
    }

    var bodyData = strings.ReplaceAll(cmd.Body, e.SaltFirst, "")
    bodyData = strings.ReplaceAll(bodyData, e.SaltAfter, "")
    bytes, err := b64.StdEncoding.DecodeString(bodyData)

	//get account
    var accountInfo = models.ForgotPasswordObject{}
    json.Unmarshal(bytes, &accountInfo)
    //fmt.Println("loginInfo = " + string(loginInfo))
    fmt.Println("account = " + accountInfo.Account)
    // Create the database handle, confirm driver is present
    db, err := sql.Open("mysql", db.Cfg.FormatDSN())
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
    var  times string

    if(rows.Next()) {
        err := rows.Scan(&id, &accounts, &password, &role, &times)
        if err != nil {
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusInternalServerError, cmd)
           return
        }

        //create token
        var newToken = myTool.RandStringBytes(16)
        fmt.Println("create new token = " + newToken)
        //save token
        var url = "https://x-space.cloud/asset/html/reset_password.html?account=" + accountInfo.Account + "&token=" + newToken
        sql := ("SELECT * FROM reset_tickets WHERE account = '" + accountInfo.Account + "';")
        tickets_row, err := db.Query(sql)
        if err != nil {
           fmt.Println("query error = "+ err.Error())
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusInternalServerError, cmd)
           return
        }
        defer tickets_row.Close()
        dt := time.Now().Format("2006-01-02 15:04:05")
        fmt.Println("dt = ", dt)
        //check has token before?
        if(tickets_row.Next()) {
          fmt.Println("old one to forgot")
          //delete old token
           _, err := db.Exec("UPDATE reset_tickets SET token_hash='" + myTool.ToMD5(newToken) + "', time='" + dt + "', token_used=0 WHERE account = '" + accountInfo.Account + "';")
           if err != nil {
             fmt.Println("UPDATE token error = " + err.Error())
             cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred (DataBase) ! \"}")
             cmd.Sign = myTool.GetSign(cmd)
             appG.Response(http.StatusOK, cmd)
             return
           }
        }else{
           //save new token to db
           _, err1 := db.Exec("INSERT INTO reset_tickets (account, token_hash, time, token_used) VALUES (?, ?, ?, ?)", accountInfo.Account, myTool.ToMD5(newToken), dt,  0)
           if err1 != nil {
              fmt.Println("add new  token error = " + err1.Error())
              cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred (DataBase) ! \"}")
              cmd.Sign = myTool.GetSign(cmd)
              appG.Response(http.StatusOK, cmd)
              return
           }
        }

        //feedback http request
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"The reset password link is sent to your mail !\"}")
        cmd.Sign = myTool.GetSign(cmd)
        appG.Response(http.StatusOK, cmd)
        //send reset password link
        go myMail.SendMail(accountInfo.Account, url)
    }else{
       //account not exist
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Account is not exist !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
    }
}

//Reset password
func ResetPassword(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
	err := c.BindJSON(&cmd)

	if err != nil {
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
		return
	}

    var sign = myTool.GetSign(cmd)

    if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
    }else{
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
       return
    }

    var bodyData = strings.ReplaceAll(cmd.Body, e.SaltFirst, "")
    bodyData = strings.ReplaceAll(bodyData, e.SaltAfter, "")
    bytes, err := b64.StdEncoding.DecodeString(bodyData)

	//get reset info
    var resetInfo = models.ResetPasswordObject{}
    json.Unmarshal(bytes, &resetInfo)

    // Create the database handle, confirm driver is present
    db, err := sql.Open("mysql", db.Cfg.FormatDSN())
    if(err != nil){
        fmt.Println("Connect to DB Failed !")
    }
    defer db.Close()

	sql := fmt.Sprintf("SELECT * FROM users WHERE account = '" + resetInfo.Account + "';")
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
    var  times string

    //check account is exits in table users
    if(rows.Next()) {
        err := rows.Scan(&id, &accounts, &password, &role, &times)
        if err != nil {
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusInternalServerError, cmd)
           return
        }

        sql := ("SELECT * FROM reset_tickets WHERE account = '" + resetInfo.Account + "';")
        tickets_row, err := db.Query(sql)
        if err != nil {
           fmt.Println("query error = "+ err.Error())
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusInternalServerError, cmd)
           return
        }
        defer tickets_row.Close()

        var  id int
        var  accounts string
        var  token_hash string
        var  expireTime string
        var  token_used int

        //check has token before?
        if(tickets_row.Next()) {
          err := tickets_row.Scan(&id, &accounts, &token_hash, &expireTime, &token_used)
          fmt.Println("query accounts = ", accounts, "token_hash = ", token_hash, "expireTime = ", expireTime , " token_used = ", token_used)
          if err != nil {
           fmt.Println("query error = ", err.Error())
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusInternalServerError, cmd)
           return
         }
         fmt.Println("https accounts = ", resetInfo.Account, " token_hash = ", resetInfo.Token, " password = ", resetInfo.Password)

         //check token
         if(strings.Compare( myTool.ToMD5(resetInfo.Token), token_hash) == 0){
             fmt.Println("token is ok !!")
             tokenExpireTime, errParse := time.ParseInLocation("2006-01-02 15:04:05", expireTime, time.Local)

             if errParse != nil {
                fmt.Println("parse time error ", errParse.Error())
                cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}")
                cmd.Sign = myTool.GetSign(cmd)
                appG.Response(http.StatusOK, cmd)
                return
             }

             dt := time.Now()
             elapsed := dt.Sub(tokenExpireTime)
             h, _ := time.ParseDuration(myTool.ShortDur(elapsed))
             //900 seconds, 15 mins
             if(h.Seconds() > 900){
                 fmt.Println("over 15 mins")
                 cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"This Link has expired !\"}")
                 cmd.Sign = myTool.GetSign(cmd)
                 appG.Response(http.StatusOK, cmd)
             }else{
                 fmt.Println("less than 15 mins")
                 if( token_used > 0 ){
                     fmt.Println("this token is used")
                     cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"This Link has expired !\"}")
                     cmd.Sign = myTool.GetSign(cmd)
                     appG.Response(http.StatusOK, cmd)
                 }else{
                    fmt.Println("this token isn't used")
                    //update password
                    _, err := db.Exec("UPDATE users SET password='" + myTool.ToMD5(resetInfo.Password) + "' WHERE account = '" + resetInfo.Account + "';")
                    if err != nil {
                       fmt.Println("UPDATE password error = " + err.Error())
                       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred (DataBase) ! \"}")
                       cmd.Sign = myTool.GetSign(cmd)
                       appG.Response(http.StatusOK, cmd)
                       return
                    }

                    _, err1 := db.Exec("UPDATE reset_tickets SET token_used=1 WHERE account = '" + resetInfo.Account + "';")
                    if err1 != nil {
                       fmt.Println("UPDATE password error = " + err1.Error())
                       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred (DataBase) ! \"}")
                       cmd.Sign = myTool.GetSign(cmd)
                       appG.Response(http.StatusOK, cmd)
                       return
                    }
                    //feedback http request
                    cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Reset password successful !\"}")
                    cmd.Sign = myTool.GetSign(cmd)
                    appG.Response(http.StatusOK, cmd)
                 }
             }
         }else{
            cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"This Link has expired !\"}")
            cmd.Sign = myTool.GetSign(cmd)
            appG.Response(http.StatusOK, cmd)
         }

        }else{
           //can't find in the reset_tickets table
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"This Link has expired !\"}")
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusOK, cmd)
        }
    }else{
       //account not exist
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Account is not exist !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
    }
}

/**
*   Get Device List
*/
func GetDeviceList(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
    var refreshToken models.RefreshTokenObject

    //var td *jwt.Todo
    tokenAuth, err := myJwt.ExtractTokenMetadata(c.Request)
    if err != nil {
        fmt.Println("GetDeviceList - need to refresh token")
        err := c.BindJSON(&refreshToken)
        if(err != nil){
          fmt.Println("GetDeviceList - error 1")
          cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Get Device error 1 !\"}")
          cmd.Sign = myTool.GetSign(cmd)
          appG.Response(http.StatusOK, cmd)
          return
        }
        fmt.Println("GetDeviceList - refresh token  = ", refreshToken.RefreshToken)
        //get new token
        var  tokenGroup = myJwt.Refresh_token(refreshToken.RefreshToken)
        if(tokenGroup != nil){
           fmt.Println("GetDeviceList - error 2 - feedback new token")
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Get Device error 2 !\" }")
           cmd.Extra = "{ \"access_token\": \"" + tokenGroup.AccessToken + "\" ,  \"refresh_token\": \"" + tokenGroup.RefreshToken + "\"}"
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusOK, cmd)
        }else{
           fmt.Println("GetDeviceList - error 3")
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Need to log-out !\"}")
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusOK, cmd)
           return
        }
    }else{
       userId, err := myJwt.FetchAuth(tokenAuth)
       if err != nil {
          fmt.Println("GetDeviceList - unauthorized !")
          cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" unauthorized !\"}")
          cmd.Sign = myTool.GetSign(cmd)
          appG.Response(http.StatusOK, cmd)
          return
       }

      var device_list = ""

      for key, _ := range passervice.GetWSList() {
         device_list += " {\"mac\": \"" + key +  "\",\"type\": \"1\",\"name\": \"2\"} ,"
      }

      device_list = myTool.RemoveLastRune(device_list)
      fmt.Println("GetDeviceList - device_list ", "{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Get device list !\", \"device_list\":[" + device_list + "]}")
      fmt.Println("GetDeviceList - userId ", userId)
      cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Get device list !\", \"device_list\":[" + device_list + "]}")
      cmd.Sign = myTool.GetSign(cmd)
      appG.Response(http.StatusOK, cmd)
    }
}

/**
* Refresh Token
*
* Use refresh token to refresh access token
*
*/
func Refresh_token(c *gin.Context){

   appG := app.Gin{C: c}
   var cmd models.Command
   var refreshToken models.RefreshTokenObject
   err := c.BindJSON(&refreshToken)
   if(err != nil){
      fmt.Println("Refresh Token error 1")
      cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Refresh Token error 1 !\"}")
      cmd.Sign = myTool.GetSign(cmd)
      appG.Response(http.StatusOK, cmd)
   }

   fmt.Println("refresh token  = ", refreshToken.RefreshToken)
   //verify the token
   os.Setenv("REFRESH_SECRET", "mcmvmkmsdnfsdmfdsjf") //this should be in an env file
   token, err := jwt.Parse(refreshToken.RefreshToken, func(token *jwt.Token) (interface{}, error) {
     //Make sure that the token method conform to "SigningMethodHMAC"
     if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
     }
     return []byte(os.Getenv("REFRESH_SECRET")), nil
   })

   //if there is an error, the token must have expired
   if err != nil {
         fmt.Println("Refresh Token error 2")
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Refresh Token error 2 !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
     return
   }

  //is token valid?
   if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
         fmt.Println("Refresh Token error 3")
      cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Refresh Token error 3 !\"}")
      cmd.Sign = myTool.GetSign(cmd)
      appG.Response(http.StatusOK, cmd)
     return
   }

   //Since token is valid, get the uuid:
   claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
   if ok && token.Valid {
     refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string
     if !ok {
           fmt.Println("Refresh Token error 4")
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Refresh Token error 4 !\"}")
        cmd.Sign = myTool.GetSign(cmd)
        appG.Response(http.StatusOK, cmd)
        return
     }
     userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
     if err != nil {
           fmt.Println("Refresh Token error 5")
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Refresh Token error 5 !\"}")
        cmd.Sign = myTool.GetSign(cmd)
        appG.Response(http.StatusOK, cmd)
        return
     }
     fmt.Println("refreshUuid ", refreshUuid)
     //Delete the previous Refresh Token
     deleted, delErr := myJwt.DeleteAuth(refreshUuid)
     if delErr != nil || deleted == 0 { //if any goes wrong
        fmt.Println("Refresh Token error 6")
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Refresh Token error 6 !\"}")
        cmd.Sign = myTool.GetSign(cmd)
        appG.Response(http.StatusOK, cmd)
        return
     }

    //Create new pairs of refresh and access tokens
     ts, createErr := myJwt.CreateToken(userId)
     if  createErr != nil {
       fmt.Println("Refresh Token error 7")
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Refresh Token error 7 !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
       return
     }

     //save the tokens metadata to redis
     saveErr := myJwt.CreateAuth(userId, ts)
     if saveErr != nil {
        fmt.Println("Refresh Token error 8")
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Refresh Token error 8 !\"}")
        cmd.Sign = myTool.GetSign(cmd)
        appG.Response(http.StatusOK, cmd)
        return
     }
 /*    tokens := map[string]string{
       "access_token":  ts.AccessToken,
       "refresh_token": ts.RefreshToken,
     }*/
     fmt.Println("access_token  = ", ts.AccessToken, " refresh_token = ", ts.RefreshToken)
     cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"access_token\": \"" + ts.AccessToken + "\" ,  \"refresh_token\": \"" + ts.RefreshToken + "\"}")
     cmd.Sign = myTool.GetSign(cmd)
     appG.Response(http.StatusOK, cmd)
   } else {
         fmt.Println("Refresh Token error 9")
      cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Refresh Token error 9 !\"}")
      cmd.Sign = myTool.GetSign(cmd)
      appG.Response(http.StatusOK, cmd)
   }
}

/**
* Logout account
*
*/
func Logout_account(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
    var refreshToken models.RefreshTokenObject
    err := c.BindJSON(&refreshToken)
    if(err != nil){
       fmt.Println("Logout error 1")
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Logout error 1 !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
       return
    }
    fmt.Println("refresh token  = ", refreshToken.RefreshToken)

    //var td *jwt.Todo
    tokenAuth, err := myJwt.ExtractTokenMetadata(c.Request)
    if err != nil {
       fmt.Println("logout refresh new token")
       //get new token
       var  tokenGroup = myJwt.Refresh_token(refreshToken.RefreshToken)
       if(tokenGroup != nil){
          //Delete the access token
          deleted, delErr := myJwt.DeleteAuth(tokenGroup.AccessUuid)
          if delErr != nil || deleted == 0 { //if any goes wrong
             fmt.Println("Logout error 2")
             cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Logout error 2 !\"}")
             cmd.Sign = myTool.GetSign(cmd)
             appG.Response(http.StatusOK, cmd)
             return
          }
          cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \" Logout successful !\" }")
          cmd.Sign = myTool.GetSign(cmd)
          appG.Response(http.StatusOK, cmd)
       }else{
          fmt.Println("Logout error 4")
          cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Logout error 4 !\"}")
          cmd.Sign = myTool.GetSign(cmd)
          appG.Response(http.StatusOK, cmd)
          return
       }
     }else{
        fmt.Println("logout not need to refresh token")
        //Delete the access token
        deleted, delErr := myJwt.DeleteAuth(tokenAuth.AccessUuid)
        if delErr != nil || deleted == 0 { //if any goes wrong
          fmt.Println("Logout error 2")
          cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Logout error 3 !\"}")
          cmd.Sign = myTool.GetSign(cmd)
          appG.Response(http.StatusOK, cmd)
          return
        }
        fmt.Println("logout successful !")
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \" Logout successful !\" }")
        cmd.Sign = myTool.GetSign(cmd)
        appG.Response(http.StatusOK, cmd)
     }
}