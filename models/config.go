package models

import (
	"encoding/json"
	"fmt"
	"time"
)

//UUID UUID v4
type UUID string

//HookConfig configuration for hooks
type HookConfig struct {
	Active      bool                       `json:"active"`
	Deleted     bool                       `json:"deleted"`
	IndexFields bool                       `json:"indexFields"`
	ID          UUID                       `json:"id"`
	Name        string                     `json:"name"`
	Type        string                     `json:"type"`
	URLContext  string                     `json:"urlContext"`
	Metas       map[string]json.RawMessage `json:"metas"`
	Date        time.Time                  `json:"-"`
}

//Validate validates required values for a hook
func (hc *HookConfig) Validate() error {
	if len(hc.Type) == 0 {
		return fmt.Errorf("type is required")
	}
	if len(hc.URLContext) == 0 {
		return fmt.Errorf("url context is required")
	}
	if len(hc.ID) == 0 {
		return fmt.Errorf("ID is required")
	}
	return nil
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
