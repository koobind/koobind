package token

import "github.com/spf13/cobra"

func init() {
	TokenCmd.AddCommand(tokenListCmd)
	TokenCmd.AddCommand(tokenDeleteCmd)
}

var TokenCmd = &cobra.Command{
	Use:     "token",
	Short:   "Manage tokens",
	Aliases: []string{"tokens"},
}
