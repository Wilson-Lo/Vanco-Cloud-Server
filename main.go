package main

import (
	"app/routers"
)

func main() {
	router := routers.InitRouter()
	router.Run(":80")
}
