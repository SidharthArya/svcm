//go:build darwin

package core

import "fmt"

type DarwinManager struct{}

func newPlatformManager(privileged bool) (ServiceManager, error) {
	return &DarwinManager{}, nil
}

func (m *DarwinManager) ListServices() ([]ServiceUnit, error) {
	return nil, fmt.Errorf("macos support not implemented yet")
}

func (m *DarwinManager) StartService(name string) error {
	return fmt.Errorf("macos support not implemented yet")
}

func (m *DarwinManager) StopService(name string) error {
	return fmt.Errorf("macos support not implemented yet")
}

func (m *DarwinManager) RestartService(name string) error {
	return fmt.Errorf("macos support not implemented yet")
}

func (m *DarwinManager) GetServiceDetails(name string) (*ServiceDetails, error) {
	return nil, fmt.Errorf("macos support not implemented yet")
}

func (m *DarwinManager) GetLogs(name string, lines int) (string, error) {
	return "", fmt.Errorf("macos support not implemented yet")
}

func (m *DarwinManager) Close() {}
