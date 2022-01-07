package passervice

import (
	"net/http"
	"app/models"
	"app/pkg/app"
	"log"
	"github.com/gorilla/websocket"
)

var wsMap map[string]*websocket.Conn
var ginMap map[string]app.Gin
var chanMap map[string]chan bool
// var LoginMap map[string]models.Device
//var signMap map[string]string

func init() {
	wsMap = make(map[string]*websocket.Conn)
	ginMap = make(map[string]app.Gin)
	//LoginMap = make(map[string]models.Device)
	chanMap = make(map[string]chan bool)
	//signMap = make(map[string]string)
}

/*func AddToSignMap(etag string, sign string) {
	signMap[etag] = sign
}*/

func AddToChanMap(etag string, c chan bool) {
	chanMap[etag] = c
}

// func AddToEtagList(etag string, device models.Device) {
// 	LoginMap[etag] = device
// }

//add websocket
func AddToWSList(serialNum string, ws *websocket.Conn) {
	wsMap[serialNum] = ws
}

//get websocket list
func GetWSList() map[string]*websocket.Conn {
     return wsMap
}

//get websocket on-line list count
func GetWSListCount() int {
     return len(wsMap)
}

//check websocket device is on-line or not
func IsDeviceOnline(name string) bool{
     if(wsMap[name] != nil){
       return true
     }else{
       return false
     }
}




//add http
func AddToGinList(etag string, gin app.Gin) {
	ginMap[etag] = gin
}

/*
func CheckSign(etag string, sign string) bool {
	if tempSign, ok := signMap[etag]; ok {
		if tempSign == sign {
			return true
		}
	}
	return false
}*/

func SendMsgToMachine(serialNum string, msg []byte) {
	if ws, ok := wsMap[serialNum]; ok {
		ws.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}


/*func SendMsgToMachine(serialNum, msg string) {
	if ws, ok := wsMap[serialNum]; ok {
		ws.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}*/

func SendResponseToHTTPRequest(etag string, cmd models.Command) {
   log.Println("SendResponseToDevice1")
	if appG, ok := ginMap[etag]; ok {
	log.Println("SendResponseToDevice2")
		appG.Response(http.StatusOK, cmd)
		if c, ok := chanMap[etag]; ok {
        			go func() {
        				c <- true
        			}()
       }
	}
}
/*
func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}*/
