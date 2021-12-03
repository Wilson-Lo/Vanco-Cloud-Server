package routers

import (
	e "app/pkg/e"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"crypto/md5"
	"app/models"
	"net/mail"
	"net/http"
    "app/pkg/app"
	"time"
	"strings"
	"io"
	"github.com/gin-gonic/gin"
	"database/sql"
    "github.com/go-sql-driver/mysql"
	 passervice "app/service/passervice"
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
		cmd.Body = "{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}"
		appG.Response(http.StatusInternalServerError, cmd)
		return
	}

    //get new account info
	var accountInfo = models.CmdCreateAccount{}
    json.Unmarshal([]byte(string(cmd.Body)), &accountInfo)

    //Valid E-mail
    if(!validEmail(accountInfo.Account)){
       cmd.Body = "{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Email address not correct !\"}"
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
       var pwMD5 = ToMD5(accountInfo.Password)
       cmd.Body = "successful"
       dt := time.Now()
       formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
               dt.Year(), dt.Month(), dt.Day(),
               dt.Hour(), dt.Minute(), dt.Second())
     //  fmt.Println("INSERT INTO users (account, password, role, time) VALUES (\"" + accountInfo.Account + "\", \"" + accountInfo.Password + "\", 2, \""+ formatted +"\")")
       _, err := db.Exec("INSERT INTO users (account, password, role, time) VALUES (\"" + accountInfo.Account + "\", \"" + pwMD5 + "\", 2, \""+ formatted +"\")")
       if err != nil {
       		cmd.Body = "{\"result\": \"" + e.SUCCESS + "\" , \"message\": \" Add new account failed\"}"
       		appG.Response(http.StatusOK, cmd)
       		return
       }
       cmd.Body = "{\"result\": \"" + e.SUCCESS + "\" , \"message\": \" Add new account successful\"}"
       appG.Response(http.StatusOK, cmd)
    }else{
       cmd.Body = "{\"result\": \"" + e.FAILURE + "\" , \"message\": \" This E-Mail has been registered\"}"
       appG.Response(http.StatusOK, cmd)
    }
}


//login account
func LoginAccount(c *gin.Context){

    appG := app.Gin{C: c}
    var cmd models.Command
	err := c.BindJSON(&cmd)

	if err != nil {
		cmd.Body = "{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}"
		appG.Response(http.StatusInternalServerError, cmd)
		return
	}

    var allData = "body="+cmd.Body+"&etag="+cmd.Etag+"&extra="+cmd.Extra+"&method="+cmd.Method+"&time="+cmd.Time+"&to="+cmd.To
    var sign = ToMD5(allData)

    if(strings.Compare(sign, cmd.Sign) == 0){
        fmt.Println("Sign is correct !")
    }else{
       cmd.Body = "{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}"
       appG.Response(http.StatusInternalServerError, cmd)
       return
    }

    var bodyData = strings.ReplaceAll(cmd.Body, "goaamaxa", "")
    bodyData = strings.ReplaceAll(bodyData, "erwrewqweasd", "")
    fmt.Println("bodyData = " + bodyData)
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
           cmd.Body = "{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Query DB error !\"}"
           appG.Response(http.StatusInternalServerError, cmd)
           return
        }

       //check password
       var userEnterPassword = ToMD5(loginInfo.Password)
       if(strings.Compare(userEnterPassword, password) == 0){
           cmd.Body = "{\"result\": \"" + e.SUCCESS + "\" , \"message\": \" Login Successful !\"}"
           appG.Response(http.StatusOK, cmd)
       }else{
          cmd.Body = "{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Passsword not correct !\"}"
          appG.Response(http.StatusOK, cmd)
       }
    }else{
       //account not exist
       cmd.Body = "{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Account not exist !\"}"
       appG.Response(http.StatusOK, cmd)
    }
}

//Valid E-mail
func validEmail(email string) bool {
    _, err := mail.ParseAddress(email)
    return err == nil
}

//string to MD5
func ToMD5(str string) string  {
    w := md5.New()
    io.WriteString(w, str)
    md5str := fmt.Sprintf("%x", w.Sum(nil))
    return md5str
}