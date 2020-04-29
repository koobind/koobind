package cmd

import (
	"github.com/spf13/cobra"
)


var jsonOutput bool

func init() {
	rootCmd.AddCommand(describeCmd)
	describeCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "", false, "Output in JSON")
}

var describeCmd = &cobra.Command{
	Use:	"describe",
	Short:  "Describe ressources",
}

