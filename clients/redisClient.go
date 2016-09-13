package clients

import (
	"fmt"
	"log"

	"gopkg.in/redis.v4"
)

var redisClient *redis.Client

func init() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error initializing redis client: %s", err))
	}

	redisClient = client
}

func GetRedisClient() *redis.Client {
	return redisClient
}
