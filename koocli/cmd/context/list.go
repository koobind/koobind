/*
Copyright (C) 2020 Serge ALEXANDRE

# This file is part of koobind project

koobind is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

koobind is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with koobind.  If not, see <http://www.gnu.org/licenses/>.
*/
package token

import (
	"fmt"
	"github.com/koobind/koobind/koocli/cmd/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

var contextListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Display local Context configuration",
	Aliases: []string{"contexts"},
	Run: func(cmd *cobra.Command, args []string) {

		currentContext := common.Context
		contexts := internal.ListContext()
		tw := new(tabwriter.Writer)
		tw.Init(os.Stdout, 2, 4, 3, ' ', 0)
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
