package executor

type ExecutorStatus struct {
	TotalWorkers map[string]int `json:"totalWorkers"`
	FreeWorkers  map[string]int `json:"freeWorkers"`
}

func Status() *ExecutorStatus {
	status := &ExecutorStatus{
		TotalWorkers: make(map[string]int),
		FreeWorkers:  make(map[string]int),
	}

	for k := range workers {
		status.TotalWorkers[k] = totalWorkersCount(k)
		status.FreeWorkers[k] = workersCount(k)
	}

	return status
}
