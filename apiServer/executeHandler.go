package apiServer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gngeorgiev/goExecutor/executor"
)

var (
	pwd          string
	tempFilesDir string
)

func init() {
	//wd, wdError := os.Getwd()
	//if wdError != nil {
	//	log.Fatal(wdError)
	//}
	//
	//pwd = wd
	//tempFilesDir = path.Join(pwd, "tmp")
	//mkdirError := os.MkdirAll(tempFilesDir, os.ModePerm)
	//if mkdirError != nil {
	//	log.Fatal(mkdirError)
	//}
}

func executeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params executor.ExecutorParams
		bindBodyError := c.BindJSON(&params)
		if bindBodyError != nil {
			c.String(http.StatusBadRequest, bindBodyError.Error())
			return
		}

		//filename := fmt.Sprintf("%s.%s", utils.RandomString(), params.Language)
		//filePath := path.Join(tempFilesDir, filename)
		//
		//writeFileError := ioutil.WriteFile(filePath, []byte(params.Code), os.ModePerm)
		//if writeFileError != nil {
		//	c.JSON(http.StatusBadRequest, writeFileError)
		//	return
		//}

		executionResult, executionError := executor.Execute(params)

		//removeFileError := os.Remove(filePath)
		//if removeFileError != nil {
		//	log.Println(removeFileError)
		//}

		if executionError != nil {
			c.String(http.StatusInternalServerError, executionError.Error())
			return
		}

		c.JSON(http.StatusOK, executionResult)
	}
}
