package apiServer

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Listen() error {
	r := gin.Default()

	r.GET("/status", statusHandler())
	r.POST("/execute", executeHandler())

	log.Println("Api Server listening")
	return r.Run(":8090")
}
