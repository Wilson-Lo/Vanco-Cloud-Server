package routers

import (
	e "app/pkg/e"
	"encoding/json"
	"fmt"
	"crypto/md5"
	"app/models"
	"net/mail"
	"net/http"
    "app/pkg/app"
	"time"
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
	    fmt.Println("parse json failed")
		cmd.Body = "{\"result\": \"" + e.FAILURE + "\" , \"message\": \" Unexpected error occurred !\"}"
		appG.Response(http.StatusInternalServerError, cmd)
		return
	}

    cmd.To = ""
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