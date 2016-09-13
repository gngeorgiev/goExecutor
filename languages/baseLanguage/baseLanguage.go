package baseLanguage

import (
	"path"

	"github.com/gngeorgiev/goExecutor/utils"
)

type Language interface {
	GetName() string
	PrepareWorkspace(workspace string) error
	ExecuteCode(ip, port, code string) (string, error)
	CleanupWorkspace(workspace string) error
	PrepareContainerFiles(folder string) error
	GetCommand() []string
	GetPort() string
	GetDefaultImage() string
}

type BaseLanguage struct {
	name string
}

func (b BaseLanguage) GetWorkspaceFilesPath(name string) string {
	return path.Join(utils.GetWd(), "workspaces", name)
}

func (b BaseLanguage) GetFilesToCopy(name string) []string {
	workspacePath := b.GetWorkspaceFilesPath(name)
	return []string{
		path.Join(workspacePath, "node_modules"),
		path.Join(workspacePath, "package.json"),
		path.Join(workspacePath, "server.js"),
	}
}

func (b BaseLanguage) GetName() string {
	return b.name
}

func (b BaseLanguage) GetPort() string {
	return "8099/tcp"
}
