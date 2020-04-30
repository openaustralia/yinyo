package jobdispatcher

// Jobs is the interface to creating jobs
type Jobs interface {
	Create(runID string, dockerImage string, command []string, maxRunTime int64, memory int64) error
	Delete(runID string) error
}
