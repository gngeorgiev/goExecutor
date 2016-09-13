package apiServer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gngeorgiev/goExecutor/executor"
)

func executeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params executor.ExecutorParams
		bindBodyError := c.BindJSON(&params)
		if bindBodyError != nil {
			c.String(http.StatusBadRequest, bindBodyError.Error())
			return
		}

		executionResult, executionError := executor.Execute(params)

		if executionError != nil {
			c.String(http.StatusInternalServerError, executionError.Error())
			return
		}

		c.JSON(http.StatusOK, executionResult)
	}
}
