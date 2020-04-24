package cmd

import (
	"fmt"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:	"logout",
	Short:  "Clear local token",
	Run:    func(cmd *cobra.Command, args []string) {
		internal.DeleteTokenBag(context)
		fmt.Printf("Bye!\n")
	},
}
