package service

import (
	//"encoding/base64"
	"app/models"
	"log"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
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
          passervice.AddToWSList(cmd.To, ws)

	case "cmd":
	      log.Println("ready to send back to http ")
		  passervice.SendResponseToHTTPRequest("wilson", cmd)

	default:
	     log.Println("default ")
	     cmd.Time = "2021"
         passervice.SendResponseToHTTPRequest("wilson", cmd)
	}
}
