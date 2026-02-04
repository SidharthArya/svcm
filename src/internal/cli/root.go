package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "svcm",
	Short: "svcm manages systemd services for the user",
	Long:  `A lightweight systemd service manager for Wayland with CLI, GUI, and MCP interfaces.`,
}

var Privileged bool

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Privileged, "privileged", "P", false, "Use system bus instead of user bus (requires sudo/policykit)")
}
