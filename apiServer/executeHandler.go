package apiServer

import (
	"net/http"

	"log"
	"os"

	"path"

	"fmt"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/gngeorgiev/goExecutor/executor"
	"github.com/gngeorgiev/goExecutor/utils"
)

type ExecuteRequest struct {
	Code     string   `json:"code"`
	Language string   `json:"language"`
	Image    string   `json:"image"`
	Command  []string `json:"command"`
}

var (
	pwd          string
	tempFilesDir string
)

func init() {
	wd, wdError := os.Getwd()
	if wdError != nil {
		log.Fatal(wdError)
	}

	pwd = wd
	tempFilesDir = path.Join(pwd, "tmp")
	mkdirError := os.MkdirAll(tempFilesDir, os.ModePerm)
	if mkdirError != nil {
		log.Fatal(mkdirError)
	}
}

func executeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request ExecuteRequest
		bindBodyError := c.BindJSON(&request)
		if bindBodyError != nil {
			c.JSON(http.StatusBadRequest, bindBodyError)
			return
		}

		filename := fmt.Sprintf("%s.%s", utils.RandomString(), request.Language)
		filePath := path.Join(tempFilesDir, filename)

		writeFileError := ioutil.WriteFile(filePath, []byte(request.Code), os.ModePerm)
		if writeFileError != nil {
			c.JSON(http.StatusBadRequest, writeFileError)
			return
		}

		executionResult, executionError := executor.Execute(executor.ExecutorParams{
			Command: request.Command,
			File:    filePath,
			Image:   request.Image,
		})

		removeFileError := os.Remove(filePath)
		if removeFileError != nil {
			log.Println(removeFileError)
		}

		if executionError != nil {
			c.JSON(http.StatusInternalServerError, executionError)
			return
		}

		c.JSON(http.StatusOK, executionResult)
	}
}
