package core

// ServiceUnit represents a systemd service unit
type ServiceUnit struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	LoadState   string `json:"load_state"`
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
}

// ServiceDetails contains extended information for status command
type ServiceDetails struct {
	ServiceUnit
	MainPID                uint32
	FragmentPath           string
	ActiveEnterTimestamp   uint64
	InactiveEnterTimestamp uint64
}

// Manager defines the interface for interacting with system services
type Manager interface {
	ListServices() ([]ServiceUnit, error)
	StartService(name string) error
	StopService(name string) error
	RestartService(name string) error
	GetServiceDetails(name string) (*ServiceDetails, error)
}
