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
package misc

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"
)
// https://medium.com/@jcox250/password-hash-salt-using-golang-b041dc94cb72

var hash_password string

func init() {
	HashCmd.PersistentFlags().StringVar(&hash_password, "password", "", "User password")

}

var HashCmd = &cobra.Command{
	Use:	"hash",
	Short:  "Provided password hash, for use in config file",
	Run:    func(cmd *cobra.Command, args []string) {
		if hash_password == "" {
			hash_password = inputPassword( "Password:")
			password2 := inputPassword( "Confirm password:")
			if hash_password != password2 {
				fmt.Printf("Passwords did not match!\n")
				return
			}
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(hash_password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", string(hash))
	},
}


func inputPassword(prompt string) string {
	_, err := fmt.Fprint(os.Stdout, prompt)
	if err != nil {
		panic(err)
	}
	bytePassword, err2 := terminal.ReadPassword(int(syscall.Stdin))
	if err2 != nil {
		panic(err2)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\n")
	return strings.TrimSpace(string(bytePassword))
}




