//go:build linux

package core

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
)

func newPlatformManager(privileged bool) (ServiceManager, error) {
	return NewSystemdManager(privileged)
}

type SystemdManager struct {
	conn       *dbus.Conn
	privileged bool
}

func NewSystemdManager(systemMode bool) (*SystemdManager, error) {
	var conn *dbus.Conn
	var err error

	if systemMode {
		conn, err = dbus.NewSystemConnectionContext(context.Background())
	} else {
		conn, err = dbus.NewUserConnectionContext(context.Background())
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to systemd bus (system=%v): %w", systemMode, err)
	}
	return &SystemdManager{conn: conn, privileged: systemMode}, nil
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

func ensureServiceSuffix(name string) string {
	if !strings.HasSuffix(name, ".service") {
		return name + ".service"
	}
	return name
}

func (m *SystemdManager) StartService(name string) error {
	name = ensureServiceSuffix(name)
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
	name = ensureServiceSuffix(name)
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
	name = ensureServiceSuffix(name)
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

func (m *SystemdManager) GetServiceDetails(name string) (*ServiceDetails, error) {
	name = ensureServiceSuffix(name)

	// Get Unit Properties
	props, err := m.conn.GetAllPropertiesContext(context.Background(), name)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties for %s: %w", name, err)
	}

	// Helper to safely get string/uint types
	getString := func(k string) string {
		if v, ok := props[k].(string); ok {
			return v
		}
		return ""
	}
	getUint32 := func(k string) uint32 {
		if v, ok := props[k].(uint32); ok {
			return v
		}
		return 0
	}
	getUint64 := func(k string) uint64 {
		if v, ok := props[k].(uint64); ok {
			return v
		}
		return 0
	}

	details := &ServiceDetails{
		ServiceUnit: ServiceUnit{
			Name:        getString("Id"),
			Description: getString("Description"),
			LoadState:   getString("LoadState"),
			ActiveState: getString("ActiveState"),
			SubState:    getString("SubState"),
		},
		MainPID:                getUint32("MainPID"),
		FragmentPath:           getString("FragmentPath"),
		ActiveEnterTimestamp:   getUint64("ActiveEnterTimestamp"),
		InactiveEnterTimestamp: getUint64("InactiveEnterTimestamp"),
	}
	return details, nil
}

func (m *SystemdManager) GetLogs(name string, lines int) (string, error) {
	var cmdArgs []string
	if m.privileged {
		cmdArgs = []string{"-u", name, "-n", strconv.Itoa(lines), "--no-pager"}
	} else {
		cmdArgs = []string{"--user", "-u", name, "-n", strconv.Itoa(lines), "--no-pager"}
	}

	cmd := exec.Command("journalctl", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("journalctl failed: %w", err)
	}
	return string(out), nil
}
