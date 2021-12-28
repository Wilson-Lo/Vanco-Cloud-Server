package service

import (
	//"encoding/base64"
	"app/models"
	"log"
	"encoding/json"
	"fmt"
	"time"
	"database/sql"
	"github.com/gorilla/websocket"
	db "app/pkg/db"
	passervice "app/service/passervice"
)

func DecodeMsg(ws *websocket.Conn, message string) {
	var cmd models.Command
	fmt.Println("msg = " , message)
	if err := json.Unmarshal([]byte(message), &cmd); err != nil {
		fmt.Println(err)
		//logging.Info(err)
	}

	//logging.Info("===================")
	//logging.Info(cmd.To)
	//logging.Info(cmd.Method)
	//logging.Info("===================")

    switch cmd.Method {

    //add new websocket to list
	case "connect":
		  log.Println("pi3 connect to Cloud : ", cmd.To)
		  //add to wsMap list (on-line)
          passervice.AddToWSList(cmd.To, ws)
          //check db has this mac or not
	      isMACUsed, err := CheckWebsocketIsExist(cmd.To)
          if(err != nil){

          }else{
             if(!isMACUsed){
                //add new device
                _,err := AddWebsocketToDB(cmd.To, "Control Box")
                if(err != nil){
                    log.Println("add websocket error ", err.Error())
                 }else{
                    log.Println("Add complete !")
                 }
             }else{
               //update last connect time
               _,err := UpdateWebSocketLastConnTime(cmd.To)
               if(err != nil){
                 log.Println("Update time error ", err.Error())
               }
             }
          }

	case "cmd":
	      log.Println("ready to send back to http ")
		  passervice.SendResponseToHTTPRequest("wilson", cmd)


	default:
	     log.Println("default ")
	     cmd.Time = "2021"
         passervice.SendResponseToHTTPRequest("wilson", cmd)
	}
}

/**
*  To check websocket is exist in the database
*/
func CheckWebsocketIsExist(mac string) (bool, error){
    // Create the database handle, confirm driver is present
	db, err := sql.Open("mysql", db.Cfg_device.FormatDSN())
	if(err != nil){
	   fmt.Println("Connect to DB Failed !")
	   return false, err
	   defer db.Close()
	}
   // See "Important settings" section.
    db.SetConnMaxLifetime(time.Minute * 3)
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(10)

   //check email is not register before
   // fmt.Println("SELECT * FROM users WHERE account = '" + accountInfo.Account + "'")
	sql := fmt.Sprintf("SELECT * FROM device_info WHERE mac = '" + mac + "'")
    rows, err := db.Query(sql)
    if err != nil {
       fmt.Println("SQLite occur error : " + err.Error())
       return false, err
    }
    defer db.Close()

    var isMACUsed = false
    for rows.Next() {
        isMACUsed = true
    }

    return isMACUsed, nil
}

/**
*  Add websocket to DB
*/
func AddWebsocketToDB(mac string, name string) (bool, error){
   // Create the database handle, confirm driver is present
	db, err := sql.Open("mysql", db.Cfg_device.FormatDSN())
	if(err != nil){
	   fmt.Println("Connect to DB Failed !")
	   return false, err
	   defer db.Close()
	}
   // See "Important settings" section.
    db.SetConnMaxLifetime(time.Minute * 3)
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(10)

    dt := time.Now()
    formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", dt.Year(), dt.Month(), dt.Day(),dt.Hour(), dt.Minute(), dt.Second())
    _, err1 := db.Exec("INSERT INTO device_info (mac, name, time, type) VALUES (\"" + mac + "\", \"" + name + "\", \"" + formatted + "\", 1)")
    if err1 != nil {
       return false, err1
    }
    return true, nil
}

/**
*  Update websocket last connect time
*/
func UpdateWebSocketLastConnTime(mac string) (bool, error){
    // Create the database handle, confirm driver is present
	db, err := sql.Open("mysql", db.Cfg_device.FormatDSN())
	if(err != nil){
	   fmt.Println("Connect to DB Failed !")
	   return false, err
	   defer db.Close()
	}
    // See "Important settings" section.
    db.SetConnMaxLifetime(time.Minute * 3)
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(10)
    dt := time.Now()
    formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", dt.Year(), dt.Month(), dt.Day(),dt.Hour(), dt.Minute(), dt.Second())
    _, err1 := db.Exec("UPDATE device_info SET time='" + formatted + "' WHERE mac = '" + mac + "';")
    if err1 != nil {
       return false, err1
    }
    return true, nil
}



