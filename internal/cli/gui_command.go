package cli

import (
	"lsysctl/internal/gui"

	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical user interface",
	Run: func(cmd *cobra.Command, args []string) {
		gui.Run()
	},
}

func init() {
	rootCmd.AddCommand(guiCmd)
}
