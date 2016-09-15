package executor

type ExecutorStatus struct {
	TotalWorkers map[string]int `json:"totalWorkers"`
	FreeWorkers  map[string]int `json:"freeWorkers"`
	MaxWorkers   int            `json:"maxWorkers"`
}

func Status() *ExecutorStatus {
	status := &ExecutorStatus{
		TotalWorkers: make(map[string]int),
		FreeWorkers:  make(map[string]int),
		MaxWorkers:   maxContainers,
	}

	for k := range workers {
		status.TotalWorkers[k] = totalWorkersCount()
		status.FreeWorkers[k] = freeWorkers()
	}

	return status
}
