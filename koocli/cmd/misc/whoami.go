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
package misc

import (
	"fmt"
	"github.com/koobind/koobind/koocli/cmd/common"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"text/tabwriter"
)

var displayToken bool

func init() {
	WhoamiCmd.PersistentFlags().BoolVar(&displayToken, "token", false, "Display token")
}

var WhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current logged user, if any",
	Run: func(cmd *cobra.Command, args []string) {
		common.InitHttpConnection()
		tokenBag := common.RetrieveTokenBag()
		if tokenBag != nil {
			tw := new(tabwriter.Writer)
			tw.Init(os.Stdout, 2, 4, 3, ' ', 0)
			if displayToken {
				_, _ = fmt.Fprintf(tw, "USER\tID\tGROUPS\tTOKEN")
				_, _ = fmt.Fprintf(tw, "\n%s\t%s\t%s\t%s", tokenBag.Username, tokenBag.Uid, strings.Join(tokenBag.Groups, ","), tokenBag.Token)
			} else {
				_, _ = fmt.Fprintf(tw, "USER\tID\tGROUPS")
				_, _ = fmt.Fprintf(tw, "\n%s\t%s\t%s", tokenBag.Username, tokenBag.Uid, strings.Join(tokenBag.Groups, ","))
			}
			_, _ = fmt.Fprintf(tw, "\n")
			_ = tw.Flush()
		} else {
			fmt.Printf("Nobody! (Not logged)\n")
			os.Exit(3)
		}
	},
}
