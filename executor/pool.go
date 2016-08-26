package executor

import (
	"fmt"
	"log"
	"path"

	"os"

	"github.com/fsouza/go-dockerclient"
	"github.com/goanywhere/fs"
	"github.com/satori/go.uuid"
)

var containers map[string]chan string

func init() {
	containers = make(map[string]chan string)

	client := getDockerClient()
	containers, listContainersError := client.ListContainers(docker.ListContainersOptions{
		Filters: map[string][]string{"name": {tag}},
	})

	if listContainersError != nil {
		log.Println(fmt.Sprintf("Error getting initial pool of containers, will retry: %s", listContainersError.Error()))
		return
	}

	for _, c := range containers {
		if containers[c.Image] == nil {
			containers[c.Image] = make(chan string)
		}

		containers[c.Image] <- c.ID
	}

}

func getContainerFromPool(p ExecutorParams) string {
	if len(containers) == 0 {
		createMoreContainers(p)
	}
}

func createMoreContainers(p ExecutorParams) {
	operationId, folder, copyFileError := copyExecutionFileToTempDir(p.File)
	if copyFileError != nil {
		return ExecutionResult{}, copyFileError
	}

	c, createContainerError := createContainer(folder, operationId, p.Image)
	if createContainerError != nil {
		return ExecutionResult{}, createContainerError
	}

	startContainerError := startContainer(c.ID)
	if startContainerError != nil {
		return ExecutionResult{}, startContainerError
	}
}

func startContainer(id string) error {
	client := getDockerClient()
	return client.StartContainer(id, nil)
}

func createContainer(mountFolder, operationId, image string) (*docker.Container, error) {
	client := getDockerClient()
	containerName := fmt.Sprintf("%s%s", tag, operationId)

	return client.CreateContainer(docker.CreateContainerOptions{
		Name: containerName,
		Config: &docker.Config{
			Tty:        true,
			Image:      image,
			Hostname:   containerName,
			WorkingDir: workDir,
			Mounts: []docker.Mount{{
				Source:      mountFolder,
				Destination: workDir,
				RW:          true,
			}},
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{fmt.Sprintf("%s:%s", mountFolder, workDir)},
		},
	})
}

func copyExecutionFileToTempDir(file string) (id, folder string, err error) {
	id = uuid.NewV4().String()
	folder = path.Join(os.ExpandEnv("$HOME"), fmt.Sprintf("%s/%s", workDir, id))
	mkdirErr := os.MkdirAll(folder, os.ModePerm)
	if mkdirErr != nil {
		return "", "", mkdirErr
	}

	cwd, cwdErr := os.Getwd()
	if cwdErr != nil {
		return "", "", cwdErr
	}

	fullSourceFilePath := path.Join(cwd, file)
	fullDestFilePath := path.Join(folder, file)
	copyErr := fs.Copy(fullSourceFilePath, fullDestFilePath)
	if copyErr != nil {
		return "", "", copyErr
	}

	err = nil
	return
}
