package protocol

// All the types here are used in the yinyo API. So, they all will get serialised and deserialised.
// Therefore, for all types include an explicit instruction for JSON marshalling/unmarshalling.

// StartRunOptions are options that can be used when starting a run
type StartRunOptions struct {
	Output   string        `json:"output"`
	Callback Callback      `json:"callback"`
	Env      []EnvVariable `json:"env"`
}

// Callback represents what we need to know to make a particular callback request
// This is not just a string so that we could support adding headers or other special things
// in the callback request
type Callback struct {
	URL string `json:"url"`
}

// EnvVariable is the name and value of an environment variable
type EnvVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ExitData holds information about how things ran and how much resources were used
type ExitData struct {
	Build *ExitDataStage `json:"build,omitempty"`
	Run   *ExitDataStage `json:"run,omitempty"`
}

// ExitDataStage gives the exit data for a single stage
type ExitDataStage struct {
	ExitCode int   `json:"exit_code"`
	Usage    Usage `json:"usage"`
}

// Usage gives the resource usage for a single stage
type Usage struct {
	WallTime   float64 `json:"wall_time"`   // In seconds
	CPUTime    float64 `json:"cpu_time"`    // In seconds
	MaxRSS     uint64  `json:"max_rss"`     // In bytes
	NetworkIn  uint64  `json:"network_in"`  // In bytes
	NetworkOut uint64  `json:"network_out"` // In bytes
}

// Run is what you get when you create a run and what you need to update it
type Run struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}