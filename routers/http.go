package routers

import (
	//"encoding/json"
	//"fmt"
	"app/models"
	//"app/service/passervice"
	"net/http"
	"app/pkg/app"
	"time"
	"github.com/gin-gonic/gin"
	 passervice "app/service/passervice"
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

	if cmd.Method == "cmd" {
		//signKey, err := connectService.GenerateSignKey(cmd)
		//if err != nil {
		//	cmd.Body = e.FAILURE
	//	} else {
		//cmd.Body = "1.0.0"
		passervice.AddToGinList("wilson", appG)
		passervice.SendMsgToMachine("wilson", string("{\"key\":\"hi\"}"))
			//cmd.Extra = signKey
		//}
	//	appG.Response(http.StatusOK, cmd)

		ch := make(chan bool)
		passervice.AddToChanMap("wilson", ch)
		select {
		case <-ch:
				break
		case <-time.After(30 * time.Second):
			cmdRes := models.Command{}
			cmdRes.Etag = cmd.Etag
			cmdRes.Body = "Timeout"
			appG.Response(http.StatusRequestTimeout, cmdRes)
			break
		}
	}
}
