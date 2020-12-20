package models

//HookConfig configuration for hooks
type HookConfig struct {
	URLContext string
}

//SetupConfig setup configuration
type SetupConfig struct {
	ReleaseName    string
	Namespace      string
	EncryptionKey  string
	ServiceAccount string
	RootDir        string
	Runtime        Runtime
	Client         Client
}

//Runtime runtime
type Runtime string

//Runtimes supported system runtimes
type Runtimes struct {
	Kubernetes Runtime
	Linux      Runtime
	Mac        Runtime
	Windows    Runtime
}

//RuntimeEnum envs runtime
var RuntimeEnum = Runtimes{
	Kubernetes: Runtime("k8s"),
	Linux:      Runtime("lin"),
	Mac:        Runtime("mac"),
	Windows:    Runtime("win"),
}

//Client Client
type Client string

//Clients supported sClients
type Clients struct {
	CLI Client
	Web Client
	API Client
}

//ClientEnum envs Client
var ClientEnum = Clients{
	CLI: Client("cli"),
	Web: Client("web"),
	API: Client("api"),
}
