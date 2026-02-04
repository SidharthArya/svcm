package core

// NewServiceManager creates a platform-specific service manager.
// privileged: If true, connects to the system-wide service manager (requires admin/root).
func NewServiceManager(privileged bool) (ServiceManager, error) {
	return newPlatformManager(privileged)
}
