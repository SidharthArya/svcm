package cli

import (
	"log"
	"svcm/src/internal/tui"

	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the k9s-style TUI",
	Run: func(cmd *cobra.Command, args []string) {
		if err := tui.Run(Privileged); err != nil {
			log.Fatalf("TUI Error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
