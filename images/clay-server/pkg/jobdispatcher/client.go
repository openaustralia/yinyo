package jobdispatcher

// TODO: Not happy with this interface. It jumbles up the concept of job and secret
type Client interface {
	// TODO: Rename to SetupJob?
	CreateJob(namePrefix string, runToken string) (string, error)
	StartJob(runName string, dockerImage string, command []string, env map[string]string) error
	DeleteJob(runName string) error
}
