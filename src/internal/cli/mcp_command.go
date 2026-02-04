package cli

import (
	"svcm/src/internal/mcp"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run the MCP server (stdio)",
	Run: func(cmd *cobra.Command, args []string) {
		mcp.Run()
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
