package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cancelCmd)
}

var cancelCmd = &cobra.Command{
	Use:	"cancel",
	Short:  "Cancel some resources",
}

