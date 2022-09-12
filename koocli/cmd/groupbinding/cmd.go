package groupbinding

import "github.com/spf13/cobra"

func init() {
	GroupBindingCmd.AddCommand(groupBindingApplyCmd)
	GroupBindingCmd.AddCommand(groupBindingCreateCmd)
	GroupBindingCmd.AddCommand(groupBindingDeleteCmd)
	GroupBindingCmd.AddCommand(groupBindingPatchCmd)
}

var GroupBindingCmd = &cobra.Command{
	Use:     "groupbinding",
	Short:   "Manage group bindings",
	Aliases: []string{"groupbindings"},
}
