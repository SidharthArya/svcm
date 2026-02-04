package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lsysctl",
	Short: "lsysctl manages systemd services for the user",
	Long:  `A lightweight systemd service manager for Wayland with CLI, GUI, and MCP interfaces.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be defined here
}
