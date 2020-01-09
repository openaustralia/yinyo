package jobdispatcher

// Jobs is the interface to creating jobs and tokens
// TODO: Not happy with this interface. It jumbles up the concept of job and secret
type Jobs interface {
	// TODO: Rename to SetupJob?
	CreateJobAndToken(namePrefix string, runToken string) (string, error)
	StartJob(runName string, dockerImage string, command []string, maxRunTime int64) error
	DeleteJobAndToken(runName string) error
}
