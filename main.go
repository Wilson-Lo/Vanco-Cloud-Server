package main

import (
	"app/routers"
	"math/rand"
	"time"
	redis "app/pkg/redis"
)

func main() {
    rand.Seed(time.Now().UnixNano())
    redis.InitRedis()
	router := routers.InitRouter()
	router.Run(":8080")
}
