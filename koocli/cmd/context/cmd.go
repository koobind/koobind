package token

import "github.com/spf13/cobra"

func init() {
	ContextCmd.AddCommand(contextListCmd)
}

var ContextCmd = &cobra.Command{
	Use:     "context",
	Short:   "Manage contexts",
	Aliases: []string{"contexts"},
}
