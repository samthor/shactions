package local

import (
	"net/http"

	sh "github.com/samthor/shactions"
)

// Device is a device managed by Manager.
type Device interface {
	ID() string
	ActionsDevice() sh.Device
	Exec([]sh.Exec) (*sh.States, error)
	Query() (*sh.States, error)
}

// ManagerConfig configures the Manager via NewManager.
type ManagerConfig struct {
	AgentUser         func(string) (string, error) // converts Authorization header to AgentUserId
	ReportStateClient *http.Client
	SyncKey           string
	Storage           Storage
	// TODO: add report state fucking magic
}
