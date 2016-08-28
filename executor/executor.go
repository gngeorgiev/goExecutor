package executor

import (
	"path"

	"os"

	"io/ioutil"

	"log"

	"fmt"

	"time"

	"net/http"

	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gngeorgiev/goExecutor/utils"
	"github.com/goanywhere/fs"
)

const (
	workDir       = "/goExecutor"
	tag           = "goExecutor-"
	maxContainers = 50
)

type ExecutorParams struct {
	Image, File string
	Command     []string
}

type ExecutionResult struct {
	Result      string `json:"result"`
	ContainerId string `json:"containerId"`
	ExecutionId string `json:"executionId"`
}

func Execute(p ExecutorParams) (ExecutionResult, error) {
	w := getWorkerFromPool(p)

	prepareWorkspaceError := prepareWorkspaceFile(w.Workspace, p.File)
	if prepareWorkspaceError != nil {
		return ExecutionResult{}, prepareWorkspaceError
	}

	execResult, execError := execContainerCommand(w.IPAddress, w.Port, p.Command)
	if execError != nil {
		return ExecutionResult{}, execError
	}

	cleanupWorkspaceError := cleanupWorkspace(w.Workspace)
	if cleanupWorkspaceError != nil {
		return ExecutionResult{}, cleanupWorkspaceError
	}

	returnWorkerToPool(p, w)

	return ExecutionResult{
		Result:      execResult,
		ContainerId: w.ContainerId,
		ExecutionId: w.Id,
	}, nil
}

func workspaceFileName(file string) string {
	return fmt.Sprintf("code%s", path.Ext(file))
}

func cleanupWorkspace(workspace string) error {
	defer utils.TrackTime(time.Now(), "cleanupWorkspace took %s")
	log.Println("Cleaning up container workspace")
	files, readDirError := ioutil.ReadDir(workspace)
	if readDirError != nil {
		return readDirError
	}

	for _, f := range files {
		removeError := os.Remove(path.Join(workspace, f.Name()))
		if removeError != nil {
			return removeError
		}
	}

	return nil
}

func prepareWorkspaceFile(workspace, file string) error {
	defer utils.TrackTime(time.Now(), "prepareWorkspaceFile took: %s")
	log.Println("Preparing execution workspace")
	workspaceFile := workspaceFileName(file)
	fullDestFilePath := path.Join(workspace, workspaceFile)
	copyErr := fs.Copy(file, fullDestFilePath)
	if copyErr != nil {
		return copyErr
	}

	return nil
}

func execContainerCommand(address, port string, cmd []string) (string, error) {
	defer utils.TrackTime(time.Now(), "execContainerCommand took: %s")
	log.Println("Executing container command")

	body := gin.H{"command": cmd}
	bodyJson, marshalError := json.Marshal(body)
	if marshalError != nil {
		return "", marshalError
	}

	r, requestError := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://%s:%s/execute", address, port),
		strings.NewReader(string(bodyJson)),
	)

	if requestError != nil {
		return "", requestError
	}

	response, doRequestError := http.DefaultClient.Do(r)
	if doRequestError != nil {
		return "", doRequestError
	}

	defer response.Body.Close()
	b, readBodyError := ioutil.ReadAll(response.Body)
	if readBodyError != nil {
		return "", readBodyError
	}

	result := string(b)
	return result, nil
}
