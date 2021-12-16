package redis

import (
  "github.com/go-redis/redis/v8"
  "os"
  "fmt"
  //"github.com/twinj/uuid"
)

var  Client *redis.Client

/**
*  Init Redis
*/
func InitRedis() {
  fmt.Println("InitRedis !")
  //Initializing redis
  dsn := os.Getenv("REDIS_DSN")
  if len(dsn) == 0 {
     dsn = "localhost:6379"
  }

  Client = redis.NewClient(&redis.Options{
     Addr: dsn, //redis port
  })
  _, err := Client.Ping(Client.Context()).Result()
  if err != nil {
     fmt.Println("init redis failed !")
     panic(err)
  }
}