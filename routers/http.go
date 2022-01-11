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
	"bytes"
    "strconv"
	"github.com/gin-gonic/gin"
	"database/sql"
	"github.com/dgrijalva/jwt-go"
)

/**
*
*  Error HTTPs Feedback
*/
func ErrorFeedback(appG app.Gin, feedbackMessage string, logMessage string){
     var cmd models.Command
     fmt.Println(logMessage)
     cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \" " + feedbackMessage + " !\"}")
     cmd.Sign = myTool.GetSign(cmd)
     appG.Response(http.StatusOK, cmd)
}

func Connect(c *gin.Context) {

    appG := app.Gin{C: c}
    var cmd models.Command
	err := c.BindJSON(&cmd)

	if err != nil {
       cmd.Body = "fail"
       appG.Response(http.StatusInternalServerError, cmd)
       ErrorFeedback(appG, "Unexpected error occurred !", "Connect - Unexpected error occurred 1 ")
       return
    }

    //var td *jwt.Todo
    _, err = myJwt.ExtractTokenMetadata(c.Request)
    if err != nil {

       fmt.Println("Connect - refresh token  = ", cmd.Extra)
       //get new token
       var  tokenGroup = myJwt.Refresh_token(cmd.Extra)
        if(tokenGroup != nil){
           fmt.Println("Connect - error 2 ")
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Unexpected error occurred !\" }")
           cmd.Extra = "{ \"access_token\": \"" + tokenGroup.AccessToken + "\" ,  \"refresh_token\": \"" + tokenGroup.RefreshToken + "\"}"
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusOK, cmd)
        }else{
           ErrorFeedback(appG, "Need to log-out !", "Connect - error 3 !")
           return
        }
    }else{

         var sign = myTool.GetSign(cmd)

         fmt.Println("Cloud Sign = ", sign)
         fmt.Println("Web Sign = ", cmd.Sign)
         fmt.Println("etag = ", cmd.Etag)

         if(strings.Compare(sign, cmd.Sign) == 0){
               fmt.Println("Sign is correct !")

         //   if cmd.Method == "cmd" {
                //signKey, err := connectService.GenerateSignKey(cmd)
                //if err != nil {
                //	cmd.Body = e.FAILURE
               passervice.AddToGinList(cmd.Etag, appG)
               reqBodyBytes := new(bytes.Buffer)
               json.NewEncoder(reqBodyBytes).Encode(cmd)
               passervice.SendMsgToMachine(cmd.To, reqBodyBytes.Bytes())

               ch := make(chan bool)
               passervice.AddToChanMap(cmd.Etag, ch)
               select {
                  case <-ch:
                        break
                  case <-time.After(6 * time.Second):
                        ErrorFeedback(appG, "Unexpected error occurred !", "Connect - Timeout !")
                        break
               }
         //   }
         }else{
            ErrorFeedback(appG, "Unexpected error occurred !", "Connect - error 4 !")
            return
        }

    }


    //var td *jwt.Todo
   /* tokenAuth, err := myJwt.ExtractTokenMetadata(c.Request)
    if err != nil {
       ErrorFeedback(appG, "Unauthorized !", "Connect - Unauthorized 1")
       return
    }

    userId, err := myJwt.FetchAuth(tokenAuth)
    if err != nil {
       ErrorFeedback(appG, "Unauthorized !", "Connect - Unauthorized 2")
       return
    }*/

   // fmt.Println("userId " , userId)

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
       ErrorFeedback(appG, "Unexpected error occurred !", "CreateAccount - Unexpected error occurred 1 ")
       return
	 }

	 var sign = myTool.GetSign(cmd)

     //check sign value
     if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
     }else{
        ErrorFeedback(appG, "Unexpected error occurred !", "CreateAccount - Unexpected error occurred 2 ")
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
       ErrorFeedback(appG, "Email address format not correct !", "CreateAccount - Email address format not correct !")
       return
    }

    // Create the database handle, confirm driver is present
	db, err := sql.Open("mysql", db.Cfg.FormatDSN())
	if(err != nil){
	   ErrorFeedback(appG, "Unexpected error occurred !", "CreateAccount - Connect to DB Failed !")
	   return
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
            ErrorFeedback(appG, "Add new account failed !", "CreateAccount - Add new account failed !")
       		return
       }

       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Register Successful !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
    }else{
       ErrorFeedback(appG, "This E-Mail has been registered !", "CreateAccount - This E-Mail has been registered !")
       return
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
	   ErrorFeedback(appG, "Unexpected error occurred !", "LoginAccount - Unexpected error occurred 1 !")
       return
	}

    var sign = myTool.GetSign(cmd)

    if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
    }else{
       ErrorFeedback(appG, "Unexpected error occurred !", "LoginAccount - Unexpected error occurred 2 !")
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
       ErrorFeedback(appG, "Unexpected error occurred (DataBase) 1 !", "LoginAccount - Unexpected error occurred (DataBase) 1 !")
       return
    }
    defer db.Close()

	sql := fmt.Sprintf("SELECT * FROM users WHERE account = '" + loginInfo.Account + "';")
    rows, err := db.Query(sql)
    if err != nil {
        ErrorFeedback(appG, "Unexpected error occurred (DataBase) 2 !", "LoginAccount - Unexpected error occurred (DataBase) 2 !")
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
           ErrorFeedback(appG, "Unexpected error occurred (DataBase) 3 !", "LoginAccount - Unexpected error occurred (DataBase) 3 !")
           return
        }

       //check password
       var userEnterPassword = myTool.ToMD5(loginInfo.Password)
       if(strings.Compare(userEnterPassword, password) == 0){
          token, err := myJwt.CreateToken(uint64(id))
          if(err != nil){
            ErrorFeedback(appG, "Unauthorized !", "LoginAccount - Create Token Error !")
            return
          }
          saveErr := myJwt.CreateAuth(uint64(id), token)
          if saveErr != nil {
             ErrorFeedback(appG, "Unauthorized !", "LoginAccount - Save Token Error !")
             return
          }
          fmt.Println("AccessToken : ", token.AccessToken, " RefreshToken : ", token.RefreshToken)
          cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Login Successful !\" , \"access_token\": \"" + token.AccessToken + "\", \"refresh_token\": \"" + token.RefreshToken + "\"}")
          cmd.Sign = myTool.GetSign(cmd)
          appG.Response(http.StatusOK, cmd)
       }else{
          ErrorFeedback(appG, "Account or Password not correct !", "LoginAccount - Account or Password not correct !")
          return
       }
    }else{
       //account not exist
       ErrorFeedback(appG, "Account or Password not correct !", "LoginAccount - Account or Password not correct !")
       return
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
	   ErrorFeedback(appG, "Unexpected error occurred !", "ForgotPassword - error 1 !")
	   return
	}

    var sign = myTool.GetSign(cmd)

    if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
    }else{
      	ErrorFeedback(appG, "Unexpected error occurred !", "ForgotPassword - error 2 !")
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
       ErrorFeedback(appG, "Unexpected error occurred (DataBase) 1 !", "ForgotPassword - Unexpected error occurred (DataBase) 1 !")
       return
    }
    defer db.Close()

	sql := fmt.Sprintf("SELECT * FROM users WHERE account = '" + accountInfo.Account + "';")
    rows, err := db.Query(sql)
    if err != nil {
       ErrorFeedback(appG, "Unexpected error occurred (DataBase) 2 !", "ForgotPassword - Unexpected error occurred (DataBase) 2 !")
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
            ErrorFeedback(appG, "Unexpected error occurred (DataBase) 3 !", "ForgotPassword - Unexpected error occurred (DataBase) 3 !")
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
            ErrorFeedback(appG, "Unexpected error occurred (DataBase) 4 !", "ForgotPassword - Unexpected error occurred (DataBase) 4 !")
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
             ErrorFeedback(appG, "Unexpected error occurred (DataBase) 5 !", "ForgotPassword - Unexpected error occurred (DataBase) 5 !")
             return
           }
        }else{
           //save new token to db
           _, err1 := db.Exec("INSERT INTO reset_tickets (account, token_hash, time, token_used) VALUES (?, ?, ?, ?)", accountInfo.Account, myTool.ToMD5(newToken), dt,  0)
           if err1 != nil {
              ErrorFeedback(appG, "Unexpected error occurred (DataBase) 6 !", "ForgotPassword - Unexpected error occurred (DataBase) 6 !")
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
       ErrorFeedback(appG, "Account is not exist !", "ForgotPassword - Account is not exist !")
       return
    }
}

//Reset password
func ResetPassword(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
	err := c.BindJSON(&cmd)

	if err != nil {
       ErrorFeedback(appG, "Unexpected error occurred !", "ResetPassword - error 1 !")
	   return
	}

    var sign = myTool.GetSign(cmd)

    if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
    }else{
       ErrorFeedback(appG, "Unexpected error occurred !", "ResetPassword - error 2 !")
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
        ErrorFeedback(appG, "Unexpected error occurred (DataBase) 1 !", "ResetPassword - Unexpected error occurred (DataBase) 1 !")
        return
    }
    defer db.Close()

	sql := fmt.Sprintf("SELECT * FROM users WHERE account = '" + resetInfo.Account + "';")
    rows, err := db.Query(sql)
    if err != nil {
       ErrorFeedback(appG, "Unexpected error occurred (DataBase) 2 !", "ResetPassword - Unexpected error occurred (DataBase) 2 !")
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
          ErrorFeedback(appG, "Unexpected error occurred (DataBase) 3 !", "ResetPassword - Unexpected error occurred (DataBase) 3 !")
          return
        }

        sql := ("SELECT * FROM reset_tickets WHERE account = '" + resetInfo.Account + "';")
        tickets_row, err := db.Query(sql)
        if err != nil {
           ErrorFeedback(appG, "Unexpected error occurred (DataBase) 4 !", "ResetPassword - Unexpected error occurred (DataBase) 4 !")
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
           ErrorFeedback(appG, "Unexpected error occurred (DataBase) 5 !", "ResetPassword - Unexpected error occurred (DataBase) 5 !")
           return
         }
         fmt.Println("https accounts = ", resetInfo.Account, " token_hash = ", resetInfo.Token, " password = ", resetInfo.Password)

         //check token
         if(strings.Compare( myTool.ToMD5(resetInfo.Token), token_hash) == 0){
             fmt.Println("token is ok !!")
             tokenExpireTime, errParse := time.ParseInLocation("2006-01-02 15:04:05", expireTime, time.Local)

             if errParse != nil {
                ErrorFeedback(appG, "Unexpected error occurred !", "ResetPassword - error 3 !")
                return
             }

             dt := time.Now()
             elapsed := dt.Sub(tokenExpireTime)
             h, _ := time.ParseDuration(myTool.ShortDur(elapsed))
             //900 seconds, 15 mins
             if(h.Seconds() > 900){//over 15 mins
                 ErrorFeedback(appG, "This Link has expired !", "ResetPassword - This Link has expired 1 !")
                 return
             }else{
                 fmt.Println("less than 15 mins")
                 if( token_used > 0 ){
                     ErrorFeedback(appG, "This Link has expired !", "ResetPassword - This Link has expired 2 !")
                     return
                 }else{
                    fmt.Println("this token isn't used")
                    //update password
                    _, err := db.Exec("UPDATE users SET password='" + myTool.ToMD5(resetInfo.Password) + "' WHERE account = '" + resetInfo.Account + "';")
                    if err != nil {
                       ErrorFeedback(appG, "Unexpected error occurred (DataBase) 6 !", "ResetPassword - Unexpected error occurred (DataBase) 6 !")
                       return
                    }

                    _, err1 := db.Exec("UPDATE reset_tickets SET token_used=1 WHERE account = '" + resetInfo.Account + "';")
                    if err1 != nil {
                       ErrorFeedback(appG, "Unexpected error occurred (DataBase) 7 !", "ResetPassword - Unexpected error occurred (DataBase) 7 !")
                       return
                    }
                    //feedback http request
                    cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Reset password successful !\"}")
                    cmd.Sign = myTool.GetSign(cmd)
                    appG.Response(http.StatusOK, cmd)
                 }
             }
         }else{
             ErrorFeedback(appG, "This Link has expired !", "ResetPassword - This Link has expired 3 !")
             return
         }

        }else{
           //can't find in the reset_tickets table
           ErrorFeedback(appG, "This Link has expired !", "ResetPassword - This Link has expired 4 !")
        }
    }else{
       //account not exist
       ErrorFeedback(appG, "Account is not exist !", "ResetPassword - Account is not exist !")
    }
}

/**
*   Get All Device List
*/
func GetAllDeviceList(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
    var refreshToken models.RefreshTokenObject

    //var td *jwt.Todo
    tokenAuth, err := myJwt.ExtractTokenMetadata(c.Request)
    if err != nil {
        fmt.Println("GetDeviceList - need to refresh token")
        err := c.BindJSON(&refreshToken)
        if(err != nil){
          ErrorFeedback(appG, "Get All Device error 1 !", "GetAllDeviceList - Get All Device error 1 !")
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
           ErrorFeedback(appG, "Need to log-out !", "GetAllDeviceList - Get All Device error 3 !")
           return
        }
    }else{
       userId, err := myJwt.FetchAuth(tokenAuth)
       if err != nil {
          ErrorFeedback(appG, "Unauthorized !", "GetAllDeviceList - unauthorized !")
          return
       }

      // var deviceMap map[string] models.DeviceInfoObject
       // Create the database handle, confirm driver is present
       db, err := sql.Open("mysql", db.Cfg_device.FormatDSN())
       if(err != nil){
          ErrorFeedback(appG, "Get All Device error 4 !", "GetAllDeviceList - Get All Device error 4 !")
          return
       }
       defer db.Close()

       // See "Important settings" section.
       db.SetConnMaxLifetime(time.Minute * 3)
       db.SetMaxOpenConns(10)
       db.SetMaxIdleConns(10)

       sql, errDB := db.Query("SELECT * FROM device_info")
       if errDB != nil {
           ErrorFeedback(appG, "Get All Device error 5 !", "GetAllDeviceList - Get All Device error 5 !")
           return
       }

       defer sql.Close()
       var device_list = ""

       //get all device
       for sql.Next() {

         var deviceInfo models.DeviceInfoObject
         err := sql.Scan(&deviceInfo.ID, &deviceInfo.Mac, &deviceInfo.Name, &deviceInfo.Time, &deviceInfo.Type, &deviceInfo.UserId)

         if err != nil {
            ErrorFeedback(appG, "Get All Device error 6 !", "GetAllDeviceList - Get All Device error 6 !")
            return
         }

         if(passervice.IsDeviceOnline(deviceInfo.Mac)){
             device_list += " {\"mac\": \"" + deviceInfo.Mac +  "\",\"status\": true ,\"user_id\": " + strconv.Itoa(deviceInfo.UserId) +  ",\"type\": " + strconv.Itoa(deviceInfo.Type) + ",\"name\": \"" + deviceInfo.Name + "\"} ,"
         }else{
             device_list += " {\"mac\": \"" + deviceInfo.Mac +  "\",\"status\": false ,\"user_id\": " + strconv.Itoa(deviceInfo.UserId) +  ",\"type\": " + strconv.Itoa(deviceInfo.Type) + ",\"name\": \"" + deviceInfo.Name + "\"} ,"
         }
      }

      device_list = myTool.RemoveLastRune(device_list)
      //fmt.Println("GetDeviceList - device_list ", "{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Get device list !\", \"device_list\":[" + device_list + "]}")
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
      ErrorFeedback(appG, "Refresh Token error 1 !", "Refresh_token - Refresh Token error 1")
      return
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
     ErrorFeedback(appG, "Refresh Token error 2 !", "Refresh_token - Refresh Token error 2")
     return
   }

  //is token valid?
   if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
     ErrorFeedback(appG, "Refresh Token error 3 !", "Refresh_token - Refresh Token error 3")
     return
   }

   //Since token is valid, get the uuid:
   claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
   if ok && token.Valid {

     refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string
     if !ok {
        ErrorFeedback(appG, "Refresh Token error 4 !", "Refresh_token - Refresh Token error 4")
        return
     }

     userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
     if err != nil {
        ErrorFeedback(appG, "Refresh Token error 5 !", "Refresh_token - Refresh Token error 5")
        return
     }
     fmt.Println("refreshUuid ", refreshUuid)
     //Delete the previous Refresh Token
     deleted, delErr := myJwt.DeleteAuth(refreshUuid)
     if delErr != nil || deleted == 0 { //if any goes wrong
        ErrorFeedback(appG, "Refresh Token error 6 !", "Refresh_token - Refresh Token error 6")
        return
     }

    //Create new pairs of refresh and access tokens
     ts, createErr := myJwt.CreateToken(userId)
     if  createErr != nil {
       ErrorFeedback(appG, "Refresh Token error 7 !", "Refresh_token - Refresh Token error 7")
       return
     }

     //save the tokens metadata to redis
     saveErr := myJwt.CreateAuth(userId, ts)
     if saveErr != nil {
        ErrorFeedback(appG, "Refresh Token error 8 !", "Refresh_token - Refresh Token error 8")
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
      ErrorFeedback(appG, "Refresh Token error 9 !", "Refresh_token - Refresh Token error 9")
      return
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
       ErrorFeedback(appG, "Logout error 1 !", "Logout_account - Logout error 1")
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
             ErrorFeedback(appG, "Logout error 2 !", "Logout_account - Logout error 2")
             return
          }
          cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \" Logout successful !\" }")
          cmd.Sign = myTool.GetSign(cmd)
          appG.Response(http.StatusOK, cmd)
       }else{
          ErrorFeedback(appG, "Logout error 4 !", "Logout_account - Logout error 4")
          return
       }
     }else{
        fmt.Println("logout not need to refresh token")
        //Delete the access token
        deleted, delErr := myJwt.DeleteAuth(tokenAuth.AccessUuid)
        if delErr != nil || deleted == 0 { //if any goes wrong
          ErrorFeedback(appG, "Logout error 3 !", "Logout_account - Logout error 3")
          return
        }
        fmt.Println("logout successful !")
        cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \" Logout successful !\" }")
        cmd.Sign = myTool.GetSign(cmd)
        appG.Response(http.StatusOK, cmd)
     }
}

/**
* Modify Device Name
*/
func Modify_Device_Name(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
    var renameInfo models.DeviceRenameObject
    err := c.BindJSON(&renameInfo)
    if(err != nil){
        ErrorFeedback(appG, "Unexpected error occurred !", "Modify_Device_Name - error 1 !")
        return
    }

    //var td *jwt.Todo
    _, err = myJwt.ExtractTokenMetadata(c.Request)
    if err != nil {

       fmt.Println("Modify_Device_Name - refresh token  = ", renameInfo.RefreshToken)
       //get new token
       var  tokenGroup = myJwt.Refresh_token(renameInfo.RefreshToken)
        if(tokenGroup != nil){
           fmt.Println("Modify_Device_Name - error 2 ")
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Get Device error 2 !\" }")
           cmd.Extra = "{ \"access_token\": \"" + tokenGroup.AccessToken + "\" ,  \"refresh_token\": \"" + tokenGroup.RefreshToken + "\"}"
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusOK, cmd)
        }else{
           ErrorFeedback(appG, "Need to log-out !", "Modify_Device_Name - error 3 !")
           return
        }

    }else{
          // Create the database handle, confirm driver is present
         db, err := sql.Open("mysql", db.Cfg_device.FormatDSN())
      	 defer db.Close()
      	 if(err != nil){
      	   ErrorFeedback(appG, "Unexpected error occurred !", "Modify_Device_Name - connect db error !")
      	   return
      	 }

         // See "Important settings" section.
         db.SetConnMaxLifetime(time.Minute * 3)
         db.SetMaxOpenConns(10)
         db.SetMaxIdleConns(10)

         _, err1 := db.Exec("UPDATE device_info SET name='" + renameInfo.Name + "' WHERE mac = '" + renameInfo.Mac + "';")
         if err1 != nil {
            ErrorFeedback(appG, "Unexpected error occurred !", "Modify_Device_Name - update db error !")
            return
         }

         //feedback http request
         cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Rename successful !\"}")
         cmd.Sign = myTool.GetSign(cmd)
         appG.Response(http.StatusOK, cmd)
    }
}

/**
* Get User Info
*/
func Get_User_Info(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
    var refreshToken models.RefreshTokenObject
    err := c.BindJSON(&refreshToken)
    if(err != nil){
        ErrorFeedback(appG, "Unexpected error occurred !", "Get_User_Info - error 1 !")
        return
    }

    //var td *jwt.Todo
    tokenAuth, err := myJwt.ExtractTokenMetadata(c.Request)
    if err != nil {

       fmt.Println("Get_User_Info - refresh token  = ", refreshToken.RefreshToken)
       //get new token
       var  tokenGroup = myJwt.Refresh_token(refreshToken.RefreshToken)
        if(tokenGroup != nil){
           fmt.Println("Get_User_Info - error 2 ")
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Get Device error 2 !\" }")
           cmd.Extra = "{ \"access_token\": \"" + tokenGroup.AccessToken + "\" ,  \"refresh_token\": \"" + tokenGroup.RefreshToken + "\"}"
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusOK, cmd)
        }else{
           ErrorFeedback(appG, "Need to log-out !", "Get_User_Info - error 3 !")
           return
        }

    }else{

       userId, err := myJwt.FetchAuth(tokenAuth)
       fmt.Println("Get_User_Info - userId  = ", userId)
       // Create the database handle, confirm driver is present
       db, err := sql.Open("mysql", db.Cfg.FormatDSN())
       defer db.Close()
       if(err != nil){
      	 ErrorFeedback(appG, "Unexpected error occurred (DataBase) 1 !", "Get_User_Info - connect db error !")
      	 return
       }

       // See "Important settings" section.
       db.SetConnMaxLifetime(time.Minute * 3)
       db.SetMaxOpenConns(10)
       db.SetMaxIdleConns(10)

       user_row := db.QueryRow("SELECT * FROM users WHERE id = ?" , userId)
       if err != nil {
           ErrorFeedback(appG, "Unexpected error occurred (DataBase) 2 !", "Get_User_Info - Unexpected error occurred (DataBase) 2 !")
           return
       }

       var  id int
       var  accounts string
       var  password string
       var  role int
       var  times string

       err = user_row.Scan(&id, &accounts, &password, &role, &times)
       fmt.Println("Get_User_Info - find user = ", accounts)
       if err != nil {
         ErrorFeedback(appG, "Unexpected error occurred (DataBase) 3 !", "Get_User_Info - Unexpected error occurred (DataBase) 3 !")
         return
       }
       //feedback http request
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"account\": \"" + accounts + "\", \"id\": " + strconv.Itoa(id) + "}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
     }
}

/**
* Get Associate Code ( Device use only, so not need to check JWT )
*/
func Get_Associate_Code(c *gin.Context){

    c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
    c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")

    appG := app.Gin{C: c}
    var cmd models.Command
    var deviceInfo models.DeviceRenameObject
    err := c.BindJSON(&deviceInfo)
    if err != nil {
       ErrorFeedback(appG, "Unexpected error occurred !", "Get_Associate_Code - Error 1 !")
       return
    }

    fmt.Println("Device mac = ", deviceInfo.Mac)

    // Create the database handle, confirm driver is present
    db, err := sql.Open("mysql", db.Cfg_device.FormatDSN())
    defer db.Close()
    if(err != nil){
       ErrorFeedback(appG, "Unexpected error occurred !", "Get_Associate_Code - connect db error !")
       return
    }

    // See "Important settings" section.
    db.SetConnMaxLifetime(time.Minute * 3)
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(10)

    //check device is connected to cloud before
    device_row := db.QueryRow("SELECT * FROM device_info WHERE mac = ?" , deviceInfo.Mac)
    if err != nil {
       ErrorFeedback(appG, "Unexpected error occurred !", "Get_Associate_Code - Unexpected error occurred (DataBase) 2 !")
       return
    }

    var  id int
    var  mac string
    var  name string
    var  times string
    var  type1 int
    var  user_id string

    err = device_row.Scan(&id, &mac, &name, &times, &type1, &user_id)

    if(err != nil){
      fmt.Println("not find device")
      ErrorFeedback(appG, "Please help the device to connect to the Internet !", "Get_Associate_Code - Please help the device to connect to the Internet !")
      return
   }

   dt := time.Now()
   formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d%02d",
               dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second(), dt.Nanosecond())
   var associate_code = myTool.ToMD5(formatted + deviceInfo.Mac)[0:8]
   fmt.Println("associate_code = ", associate_code)

   //check device is exist in table - device_associate_code
   associate_code_row := db.QueryRow("SELECT * FROM device_associate_code WHERE mac = ?" , deviceInfo.Mac)
   if err != nil {
      ErrorFeedback(appG, "Unexpected error occurred !", "Get_Associate_Code - Unexpected error occurred (DataBase) 3 !")
      return
   }

   fmt.Println("associate_code = ", associate_code)
   var  get_associate_code string
   var  code_used int
   err1 := associate_code_row.Scan(&id, &mac, &get_associate_code, &times, &code_used)

   if err1 != nil  && err != sql.ErrNoRows {
      fmt.Println("new one for device_associate_code : ")
      _, err := db.Exec("INSERT INTO device_associate_code (mac, associate_code, time, code_used) VALUES (?, ?, ?, ?)", deviceInfo.Mac, associate_code, time.Now().Format("2006-01-02 15:04:05"),  0)
      if err != nil {
           fmt.Println("error = ", err.Error())
           ErrorFeedback(appG, "Get associate code failed !", "Get_Associate_Code - Insert new error!" )
           return
      }
   }else{
      fmt.Println("old one for device_associate_code ")
      _, err := db.Exec("UPDATE device_associate_code SET associate_code='" + associate_code + "', code_used=0 WHERE mac = '" + deviceInfo.Mac + "';")
      if err != nil {
         ErrorFeedback(appG, "Unexpected error occurred !", "Get_Associate_Code - Unexpected error occurred (DataBase) 4 !")
         return
      }
   }

   //feedback http request
   cmd.Extra = associate_code
   cmd.Sign = myTool.GetSign(cmd)
   appG.Response(http.StatusOK, cmd)
}

/**
* Add Device under user account
*/
func AddDevice(c *gin.Context){

     appG := app.Gin{C: c}
     var cmd models.Command
     var addDeviceObject models.AddDeviceObject
     err := c.BindJSON(&addDeviceObject)
     if(err != nil){
        ErrorFeedback(appG, "Unexpected error occurred !", "Add Device - error 1 !")
        return
     }

     //var td *jwt.Todo
     tokenAuth, err := myJwt.ExtractTokenMetadata(c.Request)
     if err != nil {
        fmt.Println("AddDevice - refresh token  = ", addDeviceObject.RefreshToken)
        //get new token
        var  tokenGroup = myJwt.Refresh_token(addDeviceObject.RefreshToken)
        if(tokenGroup != nil){
           fmt.Println("AddDevice - error 2 ")
           cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.FAILURE + "\" , \"message\": \"Add Device error 2 !\" }")
           cmd.Extra = "{ \"access_token\": \"" + tokenGroup.AccessToken + "\" ,  \"refresh_token\": \"" + tokenGroup.RefreshToken + "\"}"
           cmd.Sign = myTool.GetSign(cmd)
           appG.Response(http.StatusOK, cmd)
        }else{
            ErrorFeedback(appG, "Need to log-out !", "AddDevice - error 2 !")
            return
        }

    }else{

       userId, err := myJwt.FetchAuth(tokenAuth)
       if err != nil {
          ErrorFeedback(appG, "Unauthorized !", "Add Device - unauthorized !")
          return
       }

       // Create the database handle, confirm driver is present
       db, err := sql.Open("mysql", db.Cfg_device.FormatDSN())
       defer db.Close()
       if(err != nil){
         ErrorFeedback(appG, "Unexpected error occurred !", "Add Device - connect db error !")
         return
       }

       // See "Important settings" section.
       db.SetConnMaxLifetime(time.Minute * 3)
       db.SetMaxOpenConns(10)
       db.SetMaxIdleConns(10)

       var  id int
       var  mac string
       var  times string
       var  get_associate_code string
       var  code_used int

      fmt.Println("AssociateCode code = ", addDeviceObject.AssociateCode)

      //check device is exist in table - device_associate_code
      associate_code_row := db.QueryRow("SELECT * FROM device_associate_code WHERE associate_code = ?" , addDeviceObject.AssociateCode)
      if err != nil {
        ErrorFeedback(appG, "Unexpected error occurred !", "Add Device - Unexpected error occurred (DataBase) 1 !")
        return
      }

      err1 := associate_code_row.Scan(&id, &mac, &get_associate_code, &times, &code_used)
      if err1 != nil  && err != sql.ErrNoRows {
         //no one has this associate code
         ErrorFeedback(appG, "The associate code has expired !", "Add Device - No one has this associate code !" )
         return
     }

     //check code is used or not
     if(code_used == 1){
        ErrorFeedback(appG, "The associate code has expired !", "Add Device - associate code is used !" )
        return
     }

     //check time
     codeExpireTime, errParse := time.ParseInLocation("2006-01-02 15:04:05", times, time.Local)
     if errParse != nil {
        ErrorFeedback(appG, "Unexpected error occurred !", "Add Device - error 3 !")
        return
     }

     dt := time.Now()
     fmt.Println("codeExpireTime = ", codeExpireTime)
     elapsed := dt.Sub(codeExpireTime)
     h, _ := time.ParseDuration(myTool.ShortDur(elapsed))

     fmt.Println("h = ", h)
     if(h.Seconds() > 300){//over 5 mins
        ErrorFeedback(appG, "This Code has expired !", "Add Device - This Code has expired !")
        return
     }else{
        //add device to user by user id
        _, err := db.Exec("UPDATE device_info SET user_id=" + strconv.Itoa(int(userId)) + " WHERE mac = '" + mac + "';")
        if err != nil {
           ErrorFeedback(appG, "Unexpected error occurred !", "Add Device - Unexpected error occurred (DataBase) 2 !")
           return
        }

        //associate code to be used
        _, err2 := db.Exec("UPDATE device_associate_code SET code_used=1 WHERE associate_code = '" + get_associate_code + "';")
        if err2 != nil {
           ErrorFeedback(appG, "Unexpected error occurred !", "Add Device - Unexpected error occurred (DataBase) 3 !")
           return
        }

       //feedback http request
       cmd.Body = myTool.EncryptionData("{\"result\": \"" + e.SUCCESS + "\" , \"message\": \"Add Device successful !\"}")
       cmd.Sign = myTool.GetSign(cmd)
       appG.Response(http.StatusOK, cmd)
      }
   }
}