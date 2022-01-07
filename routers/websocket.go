package routers

import (
    "fmt"
    "bytes"
	"net/http"
	"strings"
    "app/models"
    "encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	myTool "app/pkg/tool"
	wsservice "app/service/wsservice"
	passervice "app/service/passervice"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WSSConnect(c *gin.Context) {

	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Websocket Error : ", err.Error())
		return
	}
	defer ws.Close()
	for {
		_, message, err := ws.ReadMessage()
		if err != nil{
		   fmt.Println("Websocket list size : ", passervice.GetWSListCount())
		    for key, websocketObject := range passervice.GetWSList() {
               if(websocketObject == ws){
                  fmt.Println(key , " remove from websocket list ")
                  delete(passervice.GetWSList(), key)
                  fmt.Println("after delete Websocket list size : ", passervice.GetWSListCount())
               }
           }
		   fmt.Println("Websocket read error = ", err.Error())
		   return
		}

       fmt.Println("wss receive message = " , string(message), "-")

	   if string(message) != "" {
            var cmd models.Command

			if strings.ToLower(string(message)) == "ping" {
			    fmt.Println("wss receive ping & send pong back")
			    cmd.Method = "pong"
                cmd.Sign = myTool.GetSign(cmd)
                reqBodyBytes := new(bytes.Buffer)
                json.NewEncoder(reqBodyBytes).Encode(cmd)
				ws.WriteMessage(websocket.TextMessage, []byte(reqBodyBytes.Bytes()))
			} else {
			    fmt.Println("wss receive something need to feedback to https")
				if err != nil {
					fmt.Println("Websocket Error : ", err)
					break
				}
				//handle websocket data by method
				wsservice.DecodeMsg(ws, string(message))
			}
		}
	}
}
