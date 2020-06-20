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
package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/koobind/koobind/koocli/cmd/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/koobind/koobind/koomgr/apis/directory/v1alpha1"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

// kubectl koo create user titi --provider crdsys --commonName "TITI" --comment "Small bird" --email "titi@cartoon.com" --passwordHash '$2a$10$dO9pDmqhwCVHkqBKdjynTONHRExZm2iDX3yzii/RUgNMt0U/wvNtG' --uid 2001

var crdUserSpec *v1alpha1.UserSpec
var uid int

func init() {
	crdUserSpec = &v1alpha1.UserSpec{
	}
	CreateUserCmd.PersistentFlags().StringVar(&Provider, "provider", "_", "")
	CreateUserCmd.PersistentFlags().BoolVar(&crdUserSpec.Disabled, "disabled", false, "")
	CreateUserCmd.PersistentFlags().StringVar(&crdUserSpec.CommonName, "commonName", "", "")
	CreateUserCmd.PersistentFlags().StringVar(&crdUserSpec.Email, "email", "", "")
	CreateUserCmd.PersistentFlags().StringVar(&crdUserSpec.Comment, "comment", "", "")
	CreateUserCmd.PersistentFlags().StringVar(&crdUserSpec.PasswordHash, "passwordHash", "", "")
	CreateUserCmd.PersistentFlags().IntVar(&uid, "uid", 0, "")
}


var CreateUserCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{},
	Short:   "Create new user (Admin)",
	Hidden:  false,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("ERROR: A username must be provided!\n")
			os.Exit(2)
		}
		InitHttpConnection()
		userName := args[0]
		token := RetrieveToken()
		if token == "" {
			token = DoLogin("", "")
		}
		if cmd.PersistentFlags().Lookup("uid").Changed {
			fmt.Printf("Set uid to %d\n", uid)
			crdUserSpec.Uid = &uid
		}
		body, err := json.Marshal(crdUserSpec)
		response, err := HttpConnection.Do("POST", fmt.Sprintf("/auth/v1/admin/%s/users/%s", Provider, userName) , &internal.HttpAuth{Token: token},  bytes.NewBuffer(body))
		if err != nil {
			panic(err)
		}
		if response.StatusCode == http.StatusCreated {
			fmt.Printf("User created sucessfully.\n")
		} else {
			PrintHttpResponseMessage(response)
		}
		if response.StatusCode != http.StatusCreated {
			os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
		}
	},
}

