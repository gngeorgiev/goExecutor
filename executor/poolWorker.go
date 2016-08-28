package executor

import (
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

const (
	ContainerEnvIdKey        = "id"
	ContainerEnvNameKey      = "name"
	ContainerEnvWorkspaceKey = "workspace"
	ContainerEnvWorkdirKey   = "workdir"
)

type PoolWorker struct {
	Id, Name, ContainerId, Image, Workspace, Workdir, Port, IPAddress string
}

func newPoolWorker(c *docker.Container) PoolWorker {
	var image, id, name, containerId, workspace, workdir, port, ipAddress string
	for _, e := range c.Config.Env {
		env := strings.Split(e, "=")
		key := env[0]
		value := env[1]
		if key == ContainerEnvIdKey {
			id = value
		} else if key == ContainerEnvNameKey {
			name = value
		} else if key == ContainerEnvWorkspaceKey {
			workspace = value
		} else if key == ContainerEnvWorkdirKey {
			workdir = value
		}
	}

	containerId = c.ID
	image = c.Image
	port = c.NetworkSettings.Ports["8099/tcp"][0].HostPort
	ipAddress = c.NetworkSettings.Gateway

	return PoolWorker{
		Id:          id,
		Name:        name,
		ContainerId: containerId,
		Image:       image,
		Workspace:   workspace,
		Workdir:     workdir,
		Port:        port,
		IPAddress:   ipAddress,
	}
}
