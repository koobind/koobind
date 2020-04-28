package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(describeCmd)
}

var describeCmd = &cobra.Command{
	Use:	"describe",
	Short:  "Describe ressources",
}

