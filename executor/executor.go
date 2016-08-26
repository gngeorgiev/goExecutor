package executor

import (
	"bytes"
	"path"

	"os"

	"io/ioutil"

	"log"

	"fmt"

	"github.com/fsouza/go-dockerclient"
	"github.com/goanywhere/fs"
)

const (
	workDir       = "/goExecutor"
	tag           = "goExecutor-"
	maxContainers = 10
)

type ExecutorParams struct {
	Image, File, Command string
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

	execResult, execError := execContainerCommand(w.ContainerId, p.Command)
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
	log.Println("Preparing execution workspace")
	workspaceFile := workspaceFileName(file)
	fullDestFilePath := path.Join(workspace, workspaceFile)
	copyErr := fs.Copy(file, fullDestFilePath)
	if copyErr != nil {
		return copyErr
	}

	return nil
}

func execContainerCommand(id, cmd string) (string, error) {
	log.Println("Executing container command")
	client := getDockerClient()
	exec, createExecError := client.CreateExec(docker.CreateExecOptions{
		AttachStdout: true,
		Tty:          true,
		Cmd:          []string{"/bin/sh", "-c", cmd},
		Container:    id,
	})

	if createExecError != nil {
		return "", createExecError
	}

	var output bytes.Buffer
	startExecError := client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: &output,
	})
	if startExecError != nil {
		return "", startExecError
	}

	return output.String(), nil
}
