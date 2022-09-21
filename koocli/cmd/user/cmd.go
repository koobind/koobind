package user

import "github.com/spf13/cobra"

func init() {
	UserCmd.AddCommand(userCreateCmd)
	UserCmd.AddCommand(userApplyCmd)
	UserCmd.AddCommand(userPatchCmd)
	UserCmd.AddCommand(userDeleteCmd)
	UserCmd.AddCommand(userDescribeCmd)
}

var UserCmd = &cobra.Command{
	Use:     "user",
	Short:   "Manage users",
	Aliases: []string{"users"},
}
