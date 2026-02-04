//go:build windows

package core

import "fmt"

type WindowsManager struct{}

func newPlatformManager(privileged bool) (ServiceManager, error) {
	return &WindowsManager{}, nil
}

func (m *WindowsManager) ListServices() ([]ServiceUnit, error) {
	return nil, fmt.Errorf("windows support not implemented yet")
}

func (m *WindowsManager) StartService(name string) error {
	return fmt.Errorf("windows support not implemented yet")
}

func (m *WindowsManager) StopService(name string) error {
	return fmt.Errorf("windows support not implemented yet")
}

func (m *WindowsManager) RestartService(name string) error {
	return fmt.Errorf("windows support not implemented yet")
}

func (m *WindowsManager) GetServiceDetails(name string) (*ServiceDetails, error) {
	return nil, fmt.Errorf("windows support not implemented yet")
}

func (m *WindowsManager) GetLogs(name string, lines int) (string, error) {
	return "", fmt.Errorf("windows support not implemented yet")
}

func (m *WindowsManager) Close() {}
