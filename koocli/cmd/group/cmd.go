package group

import "github.com/spf13/cobra"

func init() {
	GroupCmd.AddCommand(groupCreateCmd)
	GroupCmd.AddCommand(groupApplyCmd)
	GroupCmd.AddCommand(groupDeleteCmd)
	GroupCmd.AddCommand(groupPatchCmd)
}

var GroupCmd = &cobra.Command{
	Use:     "group",
	Short:   "Manage groups",
	Aliases: []string{"groups"},
}
