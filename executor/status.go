package executor

type ExecutorStatus struct {
	Workers map[string]int `json:"workers"`
}

func Status() *ExecutorStatus {
	status := &ExecutorStatus{
		Workers: make(map[string]int),
	}

	for k := range workers {
		status.Workers[k] = len(workers[k])
	}

	return status
}
