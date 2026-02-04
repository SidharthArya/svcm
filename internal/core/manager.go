package core

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
)

type SystemdManager struct {
	conn *dbus.Conn
}

func NewSystemdManager() (*SystemdManager, error) {
	conn, err := dbus.NewUserConnectionContext(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to systemd user bus: %w", err)
	}
	return &SystemdManager{conn: conn}, nil
}

func (m *SystemdManager) Close() {
	m.conn.Close()
}

func (m *SystemdManager) ListServices() ([]ServiceUnit, error) {
	// ListUnitsByPatterns(states, patterns)
	// We want all service units.
	units, err := m.conn.ListUnitsByPatternsContext(context.Background(), nil, []string{"*.service"})
	if err != nil {
		return nil, fmt.Errorf("failed to list units: %w", err)
	}

	var services []ServiceUnit
	for _, u := range units {
		services = append(services, ServiceUnit{
			Name:        u.Name,
			Description: u.Description,
			LoadState:   u.LoadState,
			ActiveState: u.ActiveState,
			SubState:    u.SubState,
		})
	}
	return services, nil
}

func (m *SystemdManager) StartService(name string) error {
	// Mode "replace" is standard
	ch := make(chan string)
	_, err := m.conn.StartUnitContext(context.Background(), name, "replace", ch)
	if err != nil {
		return fmt.Errorf("failed to start service %s: %w", name, err)
	}
	result := <-ch
	if result != "done" {
		return fmt.Errorf("start job for %s failed with result: %s", name, result)
	}
	return nil
}

func (m *SystemdManager) StopService(name string) error {
	ch := make(chan string)
	_, err := m.conn.StopUnitContext(context.Background(), name, "replace", ch)
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %w", name, err)
	}
	result := <-ch
	if result != "done" {
		return fmt.Errorf("stop job for %s failed with result: %s", name, result)
	}
	return nil
}

func (m *SystemdManager) RestartService(name string) error {
	ch := make(chan string)
	_, err := m.conn.RestartUnitContext(context.Background(), name, "replace", ch)
	if err != nil {
		return fmt.Errorf("failed to restart service %s: %w", name, err)
	}
	result := <-ch
	if result != "done" {
		return fmt.Errorf("restart job for %s failed with result: %s", name, result)
	}
	return nil
}
