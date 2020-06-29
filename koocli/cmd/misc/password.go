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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/koobind/koobind/common"
	. "github.com/koobind/koobind/koocli/cmd/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var oldPassword string
var newPassword string

func init() {
	PasswordCmd.PersistentFlags().StringVar(&oldPassword, "oldPassword", "", "Old password")
	PasswordCmd.PersistentFlags().StringVar(&newPassword, "newPassword", "", "New password")
}


var PasswordCmd = &cobra.Command{
	Use:	"password",
	Short:  "Change current user password",
	Run:    func(cmd *cobra.Command, args []string) {
		InitHttpConnection()
		token := RetrieveToken()
		if token == "" {
			_, _ = fmt.Fprintf(os.Stderr, "Not logged in currently\n")
			return
		}
		if oldPassword == "" {
			oldPassword = inputPassword("Old password:")
		}
		if newPassword == "" {
			newPassword = inputPasswordWithConfirm()
			if newPassword == "" {
				_, _ = fmt.Fprintf(os.Stderr, "Too many failure !!!\n")
				return
			}
		}
		changePasswordRequest := common.ChangePasswordRequest{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}
		body, err := json.Marshal(changePasswordRequest)
		response, err := HttpConnection.Do("POST", "/auth/v1/changePassword", &internal.HttpAuth{Token: token}, bytes.NewBuffer(body))
		if err != nil {
			panic(err)
		}
		if response.StatusCode == http.StatusOK {
			fmt.Printf("Password changed successfully.\n")
			return
		} else {
			PrintHttpResponseMessage(response)
			os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
		}
	},
}

func inputPasswordWithConfirm() string {
	for i := 0; i <3; i++ {
		if i != 0 {
			fmt.Printf("Password did not match. Please retry.\n")
		}
		newPassword1 := inputPassword("New password:")
		newPassword2 := inputPassword("Confirm ew password:")
		if newPassword1 == newPassword2 {
			return newPassword1
		}
	}
	return ""
}
