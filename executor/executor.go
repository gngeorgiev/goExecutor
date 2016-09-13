package executor

import (
	"github.com/gngeorgiev/goExecutor/languages"
	"github.com/gngeorgiev/goExecutor/languages/baseLanguage"
)

const (
	ContainerWorkDir = "/goExecutor"
	tag              = "goExecutor-"
	maxContainers    = 128
)

type ExecutorParams struct {
	Key      string `json:"code"`
	Language string `json:"language"`
	Image    string `json:"image"`
}

type ExecutionResult struct {
	Result      string `json:"result"`
	ContainerId string `json:"containerId"`
	ExecutionId string `json:"executionId"`
}

func assignDefaultParams(p *ExecutorParams, l baseLanguage.Language) ExecutorParams {
	if p.Image == "" {
		p.Image = l.GetDefaultImage()
	}

	return *p
}

func Execute(p ExecutorParams) (ExecutionResult, error) {
	language, getLanguageError := languages.GetLanguage(p.Language)
	if getLanguageError != nil {
		return ExecutionResult{}, getLanguageError
	}

	params := assignDefaultParams(&p, language)
	w := getWorkerFromPool(params, language)

	execResult, execError := language.ExecuteCode(w.IPAddress, w.Port, params.Key)
	if execError != nil {
		return ExecutionResult{}, execError
	}

	returnWorkerToPool(params, w)

	return ExecutionResult{
		Result:      execResult,
		ContainerId: w.ContainerId,
		ExecutionId: w.Id,
	}, nil
}
