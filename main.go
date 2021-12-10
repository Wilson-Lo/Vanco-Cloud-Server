package main

import (
	"app/routers"
	redis "app/pkg/redis"
)

func main() {
    redis.InitRedis()
	router := routers.InitRouter()
	router.Run(":8080")
}
