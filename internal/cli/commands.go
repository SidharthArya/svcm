package cli

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"lsysctl/internal/core"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all user services",
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := core.NewSystemdManager()
		if err != nil {
			log.Fatalf("Failed to connect to systemd: %v", err)
		}
		defer manager.Close()

		services, err := manager.ListServices()
		if err != nil {
			log.Fatalf("Failed to list services: %v", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATE\tACTIVE\tDESCRIPTION")
		for _, s := range services {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Name, s.LoadState, s.ActiveState, s.Description)
		}
		w.Flush()
	},
}

var startCmd = &cobra.Command{
	Use:   "start [service]",
	Short: "Start a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := core.NewSystemdManager()
		if err != nil {
			log.Fatalf("Failed to connect to systemd: %v", err)
		}
		defer manager.Close()

		name := args[0]
		if err := manager.StartService(name); err != nil {
			log.Fatalf("Failed to start service %s: %v", name, err)
		}
		fmt.Printf("Service %s started.\n", name)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop [service]",
	Short: "Stop a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := core.NewSystemdManager()
		if err != nil {
			log.Fatalf("Failed to connect to systemd: %v", err)
		}
		defer manager.Close()

		name := args[0]
		if err := manager.StopService(name); err != nil {
			log.Fatalf("Failed to stop service %s: %v", name, err)
		}
		fmt.Printf("Service %s stopped.\n", name)
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart [service]",
	Short: "Restart a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := core.NewSystemdManager()
		if err != nil {
			log.Fatalf("Failed to connect to systemd: %v", err)
		}
		defer manager.Close()

		name := args[0]
		if err := manager.RestartService(name); err != nil {
			log.Fatalf("Failed to restart service %s: %v", name, err)
		}
		fmt.Printf("Service %s restarted.\n", name)
	},
}
