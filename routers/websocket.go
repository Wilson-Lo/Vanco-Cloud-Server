package routers

import (
    "fmt"
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	wsservice "app/service/wsservice"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WSSConnect(c *gin.Context) {

	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Websocket Error : ", err)
		return
	}
	defer ws.Close()
	for {
		_, message, err := ws.ReadMessage()
		if err != nil{
		   fmt.Println("Websocket read error = ", err.Error())
		   return
		}

		if string(message) != "" {

			if strings.ToLower(string(message)) == "ping" {
				ws.WriteMessage(websocket.TextMessage, []byte("pong"))
			} else {
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
