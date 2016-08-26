package executor

import (
	"bytes"

	"github.com/fsouza/go-dockerclient"
)

const (
	workDir = "/goExecutor"
	tag     = "goExecutor-"
)

type ExecutorParams struct {
	Image, File, Command string
}

type ExecutionResult struct {
	Result, ContainerId, ExecutionId string
}

func Run(p ExecutorParams) (ExecutionResult, error) {

	execResult, execError := execContainerCommand(c.ID, p.Command)
	if execError != nil {
		return ExecutionResult{}, execError
	}

	return ExecutionResult{
		Result:      execResult,
		ContainerId: c.ID,
		ExecutionId: operationId,
	}, nil
}

func execContainerCommand(id, cmd string) (string, error) {
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
