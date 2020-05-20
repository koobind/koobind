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
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"os"
)


var login string
var login_password string

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.PersistentFlags().StringVarP(&login, "user", "", "", "User name")
	loginCmd.PersistentFlags().StringVarP(&login_password, "password", "", "", "User password")

}


var loginCmd = &cobra.Command{
	Use:	"login",
	Short:  "Logout and get a new token",
	Run:    func(cmd *cobra.Command, args []string) {
		initHttpConnection()
		internal.DeleteTokenBag(context)	// Logout first. Don't stay logged with old token if we are unable to login
		t := doLogin(login, login_password)
		if t == "" {
			os.Exit(3)
		}
	},
}
