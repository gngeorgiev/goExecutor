package clients

import (
	"log"

	"fmt"

	"github.com/fsouza/go-dockerclient"
)

const dockerEndpoint = "unix:///var/run/docker.sock"

var dockerClient *docker.Client

func init() {
	client, clientError := docker.NewClient(dockerEndpoint)
	if clientError != nil {
		log.Fatal(fmt.Sprintf("Error initializing docker client: %s", clientError.Error()))
	}

	dockerClient = client
}

func GetDockerClient() *docker.Client {
	return dockerClient
}
