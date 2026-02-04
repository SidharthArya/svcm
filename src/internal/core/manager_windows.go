//go:build windows

package core

import (
	"fmt"
	"os/exec"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type WindowsManager struct {
	m *mgr.Mgr
}

func newPlatformManager(privileged bool) (ServiceManager, error) {
	// Connect to SCM
	// On Windows, typical access usually requires admin rights for full control,
	// but we can try opening with basic rights if not privileged?
	// Actually, Connect() without arguments connects to local SCM.
	m, err := mgr.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to service manager: %w", err)
	}
	return &WindowsManager{m: m}, nil
}

func (m *WindowsManager) ListServices() ([]ServiceUnit, error) {
	// List all services
	names, err := m.m.ListServices()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var services []ServiceUnit
	for _, name := range names {
		// Open service to check status
		// Using OpenService with Read access
		s, err := m.m.OpenService(name)
		if err != nil {
			// Skip if we can't open (permission denied etc) or log?
			continue
		}

		status, err := s.Query()
		s.Close() // Close immediately

		if err != nil {
			continue
		}

		activeState := "inactive"
		if status.State == svc.Running {
			activeState = "active"
		} else if status.State == svc.StartPending {
			activeState = "activating"
		} else if status.State == svc.StopPending {
			activeState = "deactivating"
		}

		services = append(services, ServiceUnit{
			Name:        name,
			Description: "", // Windows API for description is separate (QueryServiceConfig2), skipping for perf
			LoadState:   "loaded",
			ActiveState: activeState,
			SubState:    fmt.Sprintf("%d", status.State),
		})
	}
	return services, nil
}

func (m *WindowsManager) StartService(name string) error {
	s, err := m.m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not open service %s: %w", name, err)
	}
	defer s.Close()

	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	return nil
}

func (m *WindowsManager) StopService(name string) error {
	s, err := m.m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not open service %s: %w", name, err)
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	// Wait logic can be added here (check status until Stopped) but returning for now
	_ = status
	return nil
}

func (m *WindowsManager) RestartService(name string) error {
	// Stop then Start
	if err := m.StopService(name); err != nil {
		return err
	}
	// TODO: Need a wait loop here really
	return m.StartService(name)
}

func (m *WindowsManager) GetServiceDetails(name string) (*ServiceDetails, error) {
	// Basic implementation
	s, err := m.m.OpenService(name)
	if err != nil {
		return nil, fmt.Errorf("could not open service %s: %w", name, err)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return nil, err
	}

	// Attempt to get description
	config, err := s.Config()
	desc := ""
	if err == nil {
		desc = config.DisplayName
	}

	return &ServiceDetails{
		ServiceUnit: ServiceUnit{
			Name:        name,
			Description: desc,
			ActiveState: fmt.Sprintf("%d", status.State),
		},
		MainPID: status.ProcessId,
	}, nil
}

func (m *WindowsManager) GetLogs(name string, lines int) (string, error) {
	// Improved Log Retrieval:
	// 1. Searches both 'System' and 'Application' logs
	// 2. Matches Source/ProviderName loosely against the service name

	psScript := fmt.Sprintf(`
        $name = "%s"
        $logs = Get-EventLog -LogName System,Application -Newest %d -ErrorAction SilentlyContinue | 
                Where-Object { $_.Source -like "*$name*" -or $_.Message -like "*$name*" } |
                Sort-Object TimeGenerated
        
        if ($logs) {
            $logs | Format-Table -Property TimeGenerated, EntryType, Message -AutoSize | Out-String -Width 150
        } else {
            Write-Output "No logs found for service '$name' in System/Application Event Logs matching source/message."
        }
    `, name, lines*2) // Fetch more initially to filter

	cmd := exec.Command("powershell", "-Command", psScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("powershell log retrieval failed: %w", err)
	}
	return string(out), nil
}

func (m *WindowsManager) Close() {
	m.m.Disconnect()
}
