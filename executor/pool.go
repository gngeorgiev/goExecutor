package executor

import (
	"fmt"
	"log"
	"path"

	"os"

	"sync"

	"time"

	"strings"

	"net/http"

	"github.com/fsouza/go-dockerclient"
	"github.com/gngeorgiev/goExecutor/clients"
	"github.com/gngeorgiev/goExecutor/languages"
	"github.com/gngeorgiev/goExecutor/languages/baseLanguage"
	"github.com/gngeorgiev/goExecutor/utils"
)

var (
	workers           map[string]chan PoolWorker
	totalWorkers      map[string]int
	poolLock          sync.Mutex
	createWorkersLock sync.Mutex
)

const (
	ContainerExecutor = "containerExecutor"
)

func init() {
	groupedWorkers, getAllWorkersError := getAllWorkers()
	if getAllWorkersError != nil {
		log.Fatal(fmt.Sprintf("Error getting initial pool of workers: %s", getAllWorkersError))
	}

	workers = make(map[string]chan PoolWorker)
	totalWorkers = make(map[string]int)
	for imageKey, workersValue := range groupedWorkers {
		workers[imageKey] = make(chan PoolWorker, len(workersValue))
		for _, worker := range workersValue {
			workers[imageKey] <- worker
		}

		totalWorkers[imageKey] = len(workersValue)
	}
}

func getAllWorkers() (map[string][]PoolWorker, error) {
	client := clients.GetDockerClient()
	containers, listContainersError := client.ListContainers(docker.ListContainersOptions{
		Filters: map[string][]string{"name": {tag}},
	})

	if listContainersError != nil {
		return nil, listContainersError
	}

	groupedWorkers := map[string][]PoolWorker{}
	for _, c := range containers {
		if groupedWorkers[c.Image] == nil {
			groupedWorkers[c.Image] = make([]PoolWorker, 0)
		}

		container, inspectError := client.InspectContainer(c.ID)
		if inspectError != nil {
			return nil, inspectError
		}

		for _, env := range container.Config.Env {
			envSplit := strings.Split(env, "=")
			if envSplit[0] == ContainerEnvLanguageKey {
				lang, getLanguageError := languages.GetLanguage(envSplit[1])
				if getLanguageError != nil {
					log.Println(fmt.Sprintf("Error while getting language from container env: %s", getLanguageError))
					continue
				}

				groupedWorkers[c.Image] = append(groupedWorkers[c.Image], newPoolWorker(container, lang))
				break
			}
		}

	}

	return groupedWorkers, nil
}

func returnWorkerToPool(p ExecutorParams, w PoolWorker) {
	workers[p.Image] <- w
}

func getWorkerFromPool(p ExecutorParams, language baseLanguage.Language) PoolWorker {
	poolLock.Lock()
	if workers[p.Image] == nil {
		workers[p.Image] = make(chan PoolWorker, maxContainers)
	}
	poolLock.Unlock()

	wc := freeWorkersForImage(p.Image)
	tWc := totalWorkersCountForImage(p.Image)
	if wc <= tWc/2 {
		newContainersCount := calculateNewContainersCount(p.Image)
		createWorkers(p, newContainersCount, language)
	}

	worker := <-workers[p.Image]
	return worker
}

func freeWorkers() int {
	freeWorkers := 0
	for _, freeWorkersChannel := range workers {
		freeWorkers += len(freeWorkersChannel)
	}

	return freeWorkers
}

func freeWorkersForImage(image string) int {
	return len(workers[image])
}

func totalWorkersCount() int {
	totalWorkersCount := 0
	for _, workers := range totalWorkers {
		totalWorkersCount += workers
	}

	return totalWorkersCount
}

func totalWorkersCountForImage(image string) int {
	return totalWorkers[image]
}

func calculateNewContainersCount(image string) int {
	currentContainersCount := totalWorkersCount()
	if currentContainersCount >= maxContainers {
		log.Println(fmt.Sprintf("Reached maximum amount of containers %d", currentContainersCount))
		return totalWorkersCountForImage(image)
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

	return newContainersCount
}

func createWorkers(p ExecutorParams, newContainersCount int, language baseLanguage.Language) {
	createWorkersLock.Lock()
	defer createWorkersLock.Unlock()

	currentContainersCount := totalWorkersCountForImage(p.Image)
	if currentContainersCount == newContainersCount {
		return
	}

	log.Println(fmt.Sprintf("Creating more workers: %d, upscaling from : %d", newContainersCount, currentContainersCount))

	var createContainerWorker func(i, tries int, wg *sync.WaitGroup)
	createContainerWorker = func(i, tries int, wg *sync.WaitGroup) {
		if tries > 3 {
			log.Println(fmt.Sprintf("Error creating container number #%d after %d tries", i, tries-1))
			return
		}

		ok := false
		defer func() {
			if !ok {
				go createContainerWorker(i, tries+1, wg)
			} else {
				wg.Done()
			}
		}()
		defer utils.TrackTime(time.Now(), "createContainerWorker took: %s")

		pullImageError := ensureImageExists(p.Image)
		if pullImageError != nil {
			log.Println(fmt.Sprintf("Error while pulling image %s,: %s", p.Image, pullImageError))
			return
		}

		newWorker, createWorkerError := createWorker(p, language)
		if createWorkerError != nil {
			log.Println(fmt.Sprintf("Error while creating a worker, retrying... %s", createWorkerError))
			return
		}

		healthCheckCount := 5
		healthCheckError := healthCheckContainer(newWorker, healthCheckCount)
		if healthCheckError != nil {
			log.Println(fmt.Sprintf("Error while bringing up worker, failed %d health checks %s", healthCheckCount, newWorker))
			removeContainerError := removeContainer(newWorker)
			if removeContainerError != nil {
				log.Println(fmt.Sprintf("Error while removing container %s", removeContainerError))
			}
			return
		}

		returnWorkerToPool(p, newWorker)
		ok = true
	}

	wg := sync.WaitGroup{}
	for i := currentContainersCount; i < newContainersCount; i++ {
		wg.Add(1)
		go createContainerWorker(i, 1, &wg)
	}

	wg.Wait()
	totalWorkers[p.Image] = newContainersCount
}

func healthCheckContainer(w PoolWorker, healthCheckCount int) error {
	var healthError error
	for i := 0; i < healthCheckCount; i++ {
		_, healthError = http.Get(fmt.Sprintf("http://%s:%s/health", w.IPAddress, w.Port))
		if healthError == nil {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	return healthError
}

func removeContainer(w PoolWorker) error {
	return clients.GetDockerClient().RemoveContainer(docker.RemoveContainerOptions{
		ID:    w.ContainerId,
		Force: true,
	})
}

func ensureImageExists(image string) error {
	client := clients.GetDockerClient()
	images, listImagesError := client.ListImages(docker.ListImagesOptions{Filter: image})
	if listImagesError != nil {
		return listImagesError
	}

	if len(images) > 0 {
		return nil
	}

	return client.PullImage(docker.PullImageOptions{Repository: image}, docker.AuthConfiguration{})
}

func createWorker(p ExecutorParams, language baseLanguage.Language) (PoolWorker, error) {
	operationId := utils.RandomString()
	folder, copyFileError := prepareContainerWorkspace(operationId)
	if copyFileError != nil {
		return PoolWorker{}, copyFileError
	}

	prepareContainerFilesError := language.PrepareContainerFiles(folder)
	if prepareContainerFilesError != nil {
		return PoolWorker{}, prepareContainerFilesError
	}

	c, createContainerError := createContainer(folder, operationId, p.Image, language)
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

	return newPoolWorker(container, language), nil
}

func inspectContainer(id string) (*docker.Container, error) {
	client := clients.GetDockerClient()
	return client.InspectContainer(id)
}

func startContainer(id string) error {
	client := clients.GetDockerClient()
	startContainerError := client.StartContainer(id, nil)
	if startContainerError != nil {
		return startContainerError
	}

	return nil
}

func createContainer(workspaceFolder, operationId, image string, language baseLanguage.Language) (*docker.Container, error) {
	client := clients.GetDockerClient()
	containerName := fmt.Sprintf("%s%s", tag, operationId)

	c, createContainerError := client.CreateContainer(docker.CreateContainerOptions{
		Name: containerName,
		Config: &docker.Config{
			Image:      image,
			Hostname:   operationId,
			WorkingDir: ContainerWorkDir,
			Mounts: []docker.Mount{{
				Source:      workspaceFolder,
				Destination: ContainerWorkDir,
				RW:          true,
			}},
			Env: []string{
				fmt.Sprintf("%s=%s", ContainerEnvNameKey, containerName),
				fmt.Sprintf("%s=%s", ContainerEnvIdKey, operationId),
				fmt.Sprintf("%s=%s", ContainerEnvWorkspaceKey, workspaceFolder),
				fmt.Sprintf("%s=%s", ContainerEnvWorkdirKey, ContainerWorkDir),
				fmt.Sprintf("%s=%s", ContainerEnvLanguageKey, language.GetName()),
			},
			Cmd: language.GetCommand(),
			ExposedPorts: map[docker.Port]struct{}{
				docker.Port(language.GetPort()): {},
			},
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{fmt.Sprintf("%s:%s", workspaceFolder, ContainerWorkDir)},
			PortBindings: map[docker.Port][]docker.PortBinding{
				docker.Port(language.GetPort()): {{HostIP: "", HostPort: ""}},
			},
			PublishAllPorts: true,
		},
	})

	if createContainerError != nil {
		return nil, createContainerError
	}

	return c, nil
}

func prepareContainerWorkspace(id string) (string, error) {
	folder := path.Join(os.ExpandEnv("$HOME"), fmt.Sprintf("%s/%s", ContainerWorkDir, id))
	mkdirErr := os.MkdirAll(folder, os.ModePerm)
	if mkdirErr != nil {
		return "", mkdirErr
	}
	//
	//containerExecutorSrc := path.Join(utils.GetWd(), ContainerExecutor, ContainerExecutor)
	//src, openFileErr := os.Open(containerExecutorSrc)
	//if openFileErr != nil {
	//	return "", openFileErr
	//}
	//defer src.Close()
	//
	//containerExecutorDest := path.Join(folder, ContainerExecutor)
	//dest, outFileErr := os.Create(containerExecutorDest)
	//if outFileErr != nil {
	//	return "", outFileErr
	//}
	//
	//_, copyErr := io.Copy(dest, src)
	//if copyErr != nil {
	//	return "", copyErr
	//}
	//
	//destCloseError := dest.Close()
	//if destCloseError != nil {
	//	return "", destCloseError
	//}
	//
	//_, chmodErr := exec.Command("chmod", "+x", containerExecutorDest).Output()
	//if chmodErr != nil {
	//	return "", chmodErr
	//}

	return folder, nil
}
