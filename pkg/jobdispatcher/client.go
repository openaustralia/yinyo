package jobdispatcher

// Jobs is the interface to creating jobs
type Jobs interface {
	Create(runName string, dockerImage string, command []string, maxRunTime int64) error
	Delete(runName string) error
}
