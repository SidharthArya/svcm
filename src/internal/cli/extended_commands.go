package cli

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"svcm/src/internal/core"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
}

// statusCmd shows detailed status using DBus properties
var statusCmd = &cobra.Command{
	Use:   "status [service]",
	Short: "Show detailed status of a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := core.NewSystemdManager(Privileged)
		if err != nil {
			log.Fatalf("Failed to connect to systemd: %v", err)
		}
		defer manager.Close()

		name := args[0]
		details, err := manager.GetServiceDetails(name)
		if err != nil {
			log.Fatalf("Failed to get status for %s: %v", name, err)
		}

		fmt.Printf("● %s - %s\n", details.Name, details.Description)
		fmt.Printf("   Loaded: %s (%s)\n", details.LoadState, details.FragmentPath)
		fmt.Printf("   Active: %s (%s)\n", details.ActiveState, details.SubState)
		if details.MainPID != 0 {
			fmt.Printf(" Main PID: %d\n", details.MainPID)
		}

		// Format Timestamps (microsecond resolution)
		if details.ActiveEnterTimestamp > 0 {
			ts := time.UnixMicro(int64(details.ActiveEnterTimestamp))
			fmt.Printf("   Active Since: %s\n", ts.Format(time.RFC1123))
		}
	},
}

// logsCmd wraps journalctl to show logs for a specific service
var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Show logs for a specific service (wrapper around journalctl)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		// Append .service if not present, though journalctl might be smart, better be explicit
		// We can reuse the core logic but here we just want to run the command
		// Actually, let's use the same name resolution logic if possible or just let journalctl handle it.
		// But the user liked smart naming.

		// Let's rely on simple string manipulation here since we are in CLI package
		if len(name) > 0 && name[len(name)-8:] != ".service" {
			name = name + ".service"
		}

		// Run journalctl -u <service> -n 50 --no-pager
		// If Privileged, run system journalctl (sudo usually required for reading system logs depending on config, but command is same structure minus --user)
		var cmdArgs []string
		if Privileged {
			cmdArgs = []string{"-u", name, "-n", "50", "--no-pager"}
		} else {
			cmdArgs = []string{"--user", "-u", name, "-n", "50", "--no-pager"}
		}

		c := exec.Command("journalctl", cmdArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			log.Fatalf("Failed to retrieve logs: %v", err)
		}
	},
}
