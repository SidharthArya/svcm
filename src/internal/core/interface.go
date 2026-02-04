package core

type ServiceManager interface {
	ListServices() ([]ServiceUnit, error)
	StartService(name string) error
	StopService(name string) error
	RestartService(name string) error
	GetServiceDetails(name string) (*ServiceDetails, error)
	GetLogs(name string, lines int) (string, error)
	Close()
}
