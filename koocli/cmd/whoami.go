/*
  Copyright (C) 2020 Serge ALEXANDRE

  This file is part of koobind project

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
package cmd

import (
	"fmt"
	. "github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var displayToken bool

func init() {
	rootCmd.AddCommand(whoamiCmd)
	whoamiCmd.PersistentFlags().BoolVarP(&displayToken, "token", "", false, "Display token")
}

var whoamiCmd = &cobra.Command{
	Use:	"whoami",
	Short:  "Display current logged user, if any",
	Run:    func(cmd *cobra.Command, args []string) {
		initHttpConnection()
		tokenBag := internal.LoadTokenBag(context)
		if user := getUser(tokenBag); user != nil {
			if displayToken {
				fmt.Printf("user:%s  id:%s  groups:%s  token:%s\n", user.Username, user.Uid, strings.Join(user.Groups, ","), tokenBag.Token)
			} else {
				fmt.Printf("user:%s  id:%s  groups:%s\n", user.Username, user.Uid, strings.Join(user.Groups, ","))
			}
		} else {
			fmt.Printf("Nobody! (Not logged)\n")
			os.Exit(3)
		}
	},
}


// getUser() trigger a server exchange (validateToken) in all cases, as we have no local storage of user info.
func getUser(tokenBag *internal.TokenBag) *User {
	if tokenBag != nil {
		user := validateToken(tokenBag.Token)
		if user != nil {
			return user
		}
	}
	return nil
}

