package main

import (
	"fmt"
	"log"

	"os"

	"os/exec"

	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	PORT = ":8099"
)

var (
	workingDirectory string
)

func init() {
	workingDirectory = os.Getenv("workdir")

	chdirError := os.Chdir(workingDirectory)
	if chdirError != nil {
		log.Fatal(fmt.Sprintf("Error while changing PWD %s", chdirError))
	}
}

type ContainerExecuteRequest struct {
	Command []string `json:"command"`
	Code    string   `json:"code"`
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	r := gin.Default()

	r.POST("/execute", func(c *gin.Context) {
		var request ContainerExecuteRequest
		c.BindJSON(&request)

		cmd := exec.Command(request.Command[0], request.Command[1:]...)
		cmdOutput, cmdError := cmd.CombinedOutput()
		if cmdError != nil {
			c.JSON(http.StatusBadRequest, cmdError)
		}

		c.String(http.StatusOK, string(cmdOutput))
	})

	log.Println(fmt.Sprintf("Container executor listening on port %s", PORT))
	log.Fatal(r.Run(PORT))
}
