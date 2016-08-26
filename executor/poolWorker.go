package executor

import (
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

const (
	ContainerEnvIdKey        = "id"
	ContainerEnvNameKey      = "name"
	ContainerEnvWorkspaceKey = "workspace"
)

type PoolWorker struct {
	Id, Name, ContainerId, Image, Workspace string
}

func newPoolWorker(c *docker.Container) PoolWorker {
	var image, id, name, containerId, workspace string
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
		}
	}

	containerId = c.ID
	image = c.Image

	return PoolWorker{
		Id:          id,
		Name:        name,
		ContainerId: containerId,
		Image:       image,
		Workspace:   workspace,
	}
}
