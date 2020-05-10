package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)



func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:	"version",
	Short:  "Display current version",
	Run:    func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version %s\n", version)
	},
}


