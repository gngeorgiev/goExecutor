package apiServer

import (
	"net/http"

	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/gngeorgiev/goExecutor/codeStorage"
)

func prepareHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, _, formError := c.Request.FormFile("code")
		if formError != nil {
			c.AbortWithError(http.StatusBadRequest, formError)
			return
		}

		key := c.Request.FormValue("key")

		data, readFileErr := ioutil.ReadAll(file)
		if readFileErr != nil {
			c.AbortWithError(http.StatusBadRequest, readFileErr)
			return
		}

		storeCodeError := codeStorage.StoreCode(key, data)
		if storeCodeError != nil {
			c.AbortWithError(http.StatusBadRequest, storeCodeError)
			return
		}

		c.Status(http.StatusOK)
	}
}
