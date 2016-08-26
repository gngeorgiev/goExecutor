package executor

import (
	"fmt"
	"log"
	"path"

	"os"

	"sync"

	"github.com/fsouza/go-dockerclient"
	"github.com/gngeorgiev/goExecutor/utils"
)

var (
	workers           map[string]chan PoolWorker
	poolLock          sync.Mutex
	resizeWorkersLock sync.Mutex
)

func init() {
	client := getDockerClient()
	containers, listContainersError := client.ListContainers(docker.ListContainersOptions{
		Filters: map[string][]string{"name": {tag}},
	})

	if listContainersError != nil {
		log.Fatal(fmt.Sprintf("Error getting initial pool of containers: %s", listContainersError))
	}

	groupedWorkers := map[string][]PoolWorker{}
	for _, c := range containers {
		if groupedWorkers[c.Image] == nil {
			groupedWorkers[c.Image] = make([]PoolWorker, 0)
		}

		container, inspectError := client.InspectContainer(c.ID)
		if inspectError != nil {
			log.Fatal(fmt.Sprintf("Error inspecting container during pool inititialization: %s", inspectError))
		}

		groupedWorkers[c.Image] = append(groupedWorkers[c.Image], newPoolWorker(container))
	}

	workers = make(map[string]chan PoolWorker)
	for imageKey, workersValue := range groupedWorkers {
		workers[imageKey] = make(chan PoolWorker, len(workersValue))
		for _, worker := range workersValue {
			workers[imageKey] <- worker
		}
	}
}

func returnWorkerToPool(p ExecutorParams, w PoolWorker) {
	log.Println("Returing the worker to the pool")
	workers[p.Image] <- w
}

func getWorkerFromPool(p ExecutorParams) PoolWorker {
	log.Println("Getting worker from pool")
	poolLock.Lock()
	log.Println(workersCount(p))
	if workers[p.Image] == nil || workersCount(p) == 0 {
		createWorkers(p)
	}
	poolLock.Unlock()
	w := <-workers[p.Image]
	log.Println(workersCount(p))
	return w
}

func workersCount(p ExecutorParams) int {
	return len(workers[p.Image])
}

func resizeWorkers(p ExecutorParams, newSize int) {
	log.Println(fmt.Sprintf("Resizing workers pool from %d to %d", workersCount(p), newSize))

	resizeWorkersLock.Lock()
	defer resizeWorkersLock.Unlock()

	newWorkers := make(chan PoolWorker, newSize)
	oldWorkers := workers[p.Image]

	if len(oldWorkers) > 0 {
		for w := range oldWorkers {
			newWorkers <- w
		}
	}

	if oldWorkers != nil {
		close(oldWorkers)
	}

	workers[p.Image] = newWorkers
}

func createWorkers(p ExecutorParams) {
	log.Println("Checking if creating new workers is neccesary")
	currentContainersCount := workersCount(p)
	if currentContainersCount >= maxContainers {
		log.Println(fmt.Sprintf("Reached maximum amount of containers %d", currentContainersCount))
		return
	}

	var newContainersCount int
	if currentContainersCount == 0 {
		newContainersCount = 1
	} else {
		newContainersCount = currentContainersCount * 2
		if newContainersCount > maxContainers {
			newContainersCount = maxContainers
		}
	}

	if newContainersCount > workersCount(p) {
		resizeWorkers(p, newContainersCount)
	}

	go func() {
		for i := currentContainersCount; i < newContainersCount; i++ {
			log.Println(fmt.Sprintf("Creating new container #%d", i))
			pullImageError := ensureImageExists(p.Image)
			if pullImageError != nil {
				log.Println(fmt.Sprintf("Error while pulling image %s,: %s", p.Image, pullImageError))
				break
			}

			newWorker, createWorkerError := createWorker(p)
			if createWorkerError != nil {
				log.Println(fmt.Sprintf("Error while creating a worker, retrying... %s", createWorkerError))
				i--
				continue
			}

			log.Println("Created a new worker")
			returnWorkerToPool(p, newWorker)
		}
	}()
}

func ensureImageExists(image string) error {
	log.Println("Ensuring that the docker image exists")
	client := getDockerClient()
	images, listImagesError := client.ListImages(docker.ListImagesOptions{Filter: image})
	if listImagesError != nil {
		return listImagesError
	}

	if len(images) > 0 {
		return nil
	}

	return client.PullImage(docker.PullImageOptions{Repository: image}, docker.AuthConfiguration{})
}

func createWorker(p ExecutorParams) (PoolWorker, error) {
	log.Println("Creating a worker")
	operationId := utils.RandomString()
	folder, copyFileError := prepareContainerWorkspace(operationId)
	if copyFileError != nil {
		return PoolWorker{}, copyFileError
	}

	c, createContainerError := createContainer(folder, operationId, p.Image)
	if createContainerError != nil {
		return PoolWorker{}, createContainerError
	}

	startContainerError := startContainer(c.ID)
	if startContainerError != nil {
		return PoolWorker{}, startContainerError
	}

	container, inspectError := inspectContainer(c.ID)
	if inspectError != nil {
		return PoolWorker{}, inspectError
	}

	return newPoolWorker(container), nil
}

func inspectContainer(id string) (*docker.Container, error) {
	log.Println("Inspecting the container")
	client := getDockerClient()
	return client.InspectContainer(id)
}

func startContainer(id string) error {
	log.Println("Starting the container")
	client := getDockerClient()
	startContainerError := client.StartContainer(id, nil)
	if startContainerError != nil {
		return startContainerError
	}

	return nil
}

func createContainer(workspaceFolder, operationId, image string) (*docker.Container, error) {
	log.Println("Creating the container")
	client := getDockerClient()
	containerName := fmt.Sprintf("%s%s", tag, operationId)

	c, createContainerError := client.CreateContainer(docker.CreateContainerOptions{
		Name: containerName,
		Config: &docker.Config{
			Tty:        true,
			Image:      image,
			Hostname:   operationId,
			WorkingDir: workDir,
			Mounts: []docker.Mount{{
				Source:      workspaceFolder,
				Destination: workDir,
				RW:          true,
			}},
			Env: []string{
				fmt.Sprintf("name=%s", containerName),
				fmt.Sprintf("id=%s", operationId),
				fmt.Sprintf("workspace=%s", workspaceFolder),
			},
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{fmt.Sprintf("%s:%s", workspaceFolder, workDir)},
		},
	})

	if createContainerError != nil {
		return nil, createContainerError
	}

	return c, nil
}

func prepareContainerWorkspace(id string) (string, error) {
	log.Println("Preparing the container workspace")
	folder := path.Join(os.ExpandEnv("$HOME"), fmt.Sprintf("%s/%s", workDir, id))
	mkdirErr := os.MkdirAll(folder, os.ModePerm)
	if mkdirErr != nil {
		return "", mkdirErr
	}

	return folder, nil
}
