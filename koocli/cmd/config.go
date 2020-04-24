package cmd

import (
	"fmt"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:	"config",
	Short:  "Display configuration",
	Run:    func(cmd *cobra.Command, args []string) {

		currentContext := context
		contexts := internal.ListContext()
		tw := new(tabwriter.Writer)
		tw.Init(os.Stdout, 2, 4, 1, ' ', 0)
		_, _ = fmt.Fprintf(tw, " \tCONTEXT\tSERVER\tCA")
		for _, ctx := range contexts {
			var mark string
			if ctx == currentContext {
				mark = "*"
			} else {
				mark = ""
			}
			myConfig := internal.LoadConfig(ctx)
			_, _ = fmt.Fprintf(tw, "\n%s\t%s\t%s\t%s", mark, ctx, myConfig.Server, myConfig.RootCaFile)
		}
		_, _ = fmt.Fprintf(tw, "\n")
		_ = tw.Flush()
		//fmt.Printf("Contexts:%v\n", contexts)
	},
}
