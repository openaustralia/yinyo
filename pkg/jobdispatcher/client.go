package jobdispatcher

// Client is the interface to creating jobs and tokens
// TODO: Not happy with this interface. It jumbles up the concept of job and secret
type Client interface {
	// TODO: Rename to SetupJob?
	CreateJobAndToken(namePrefix string, runToken string) (string, error)
	StartJob(runName string, dockerImage string, command []string) error
	DeleteJobAndToken(runName string) error
	GetToken(runName string) (string, error)
}
