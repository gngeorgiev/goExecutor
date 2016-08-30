package executor

import "github.com/gngeorgiev/goExecutor/languages"

const (
	ContainerWorkDir = "/goExecutor"
	tag              = "goExecutor-"
	maxContainers    = 128
)

type ExecutorParams struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Image    string `json:"image"`
}

type ExecutionResult struct {
	Result      string `json:"result"`
	ContainerId string `json:"containerId"`
	ExecutionId string `json:"executionId"`
}

func Execute(p ExecutorParams) (ExecutionResult, error) {
	language, getLanguageError := languages.GetLanguage(p.Language)
	if getLanguageError != nil {
		return ExecutionResult{}, getLanguageError
	}

	w := getWorkerFromPool(p, language)

	execResult, execError := language.ExecuteCode(w.IPAddress, w.Port, p.Code)
	if execError != nil {
		return ExecutionResult{}, execError
	}

	returnWorkerToPool(p, w)

	return ExecutionResult{
		Result:      execResult,
		ContainerId: w.ContainerId,
		ExecutionId: w.Id,
	}, nil
}
