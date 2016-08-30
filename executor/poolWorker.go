package executor

import (
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/gngeorgiev/goExecutor/languages/baseLanguage"
)

const (
	ContainerEnvIdKey        = "id"
	ContainerEnvNameKey      = "name"
	ContainerEnvWorkspaceKey = "workspace"
	ContainerEnvWorkdirKey   = "workdir"
	ContainerEnvLanguageKey  = "language"
)

type PoolWorker struct {
	Id, Name, ContainerId, Image, Workspace, Workdir, Port, IPAddress, Language string
}

func newPoolWorker(c *docker.Container, language baseLanguage.Language) PoolWorker {
	var image, id, name, containerId, workspace, workdir, port, ipAddress, lang string
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
		} else if key == ContainerEnvLanguageKey {
			lang = value
		}
	}

	containerId = c.ID
	image = c.Image
	port = c.NetworkSettings.Ports[docker.Port(language.GetPort())][0].HostPort
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
		Language:    lang,
	}
}
