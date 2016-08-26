package apiServer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gngeorgiev/goExecutor/executor"
)

func statusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		status := executor.Status()
		c.JSON(http.StatusOK, status)
	}
}
