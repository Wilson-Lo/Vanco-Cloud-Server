package redis

import (
  "github.com/go-redis/redis/v8"
  "os"
  "fmt"
  //"github.com/twinj/uuid"
)

var  client *redis.Client

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

  client = redis.NewClient(&redis.Options{
     Addr: dsn, //redis port
  })
  _, err := client.Ping(client.Context()).Result()
  if err != nil {
     panic(err)
  }
}