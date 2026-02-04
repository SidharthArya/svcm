package cli

import (
	"fmt"
	"log"
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
		manager, err := core.NewServiceManager(Privileged)
		if err != nil {
			log.Fatalf("Failed to connect to service manager: %v", err)
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

// logsCmd shows logs using the backend manager
var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Show logs for a specific service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := core.NewServiceManager(Privileged)
		if err != nil {
			log.Fatalf("Failed to connect to service manager: %v", err)
		}
		defer manager.Close()

		name := args[0]
		// 50 lines default
		logs, err := manager.GetLogs(name, 50)
		if err != nil {
			log.Fatalf("Failed to retrieve logs: %v", err)
		}
		fmt.Print(logs)
	},
}
