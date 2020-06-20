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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/koobind/koobind/koomgr/apis/directory/v1alpha1"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var crdUserSpec *v1alpha1.UserSpec
var uid int
var provider string

func init() {
	crdUserSpec = &v1alpha1.UserSpec{
	}
	CreateCmd.AddCommand(createUserCmd)
	createUserCmd.PersistentFlags().StringVar(&provider, "provider", "_", "")
	createUserCmd.PersistentFlags().BoolVar(&crdUserSpec.Disabled, "disabled", false, "")
	createUserCmd.PersistentFlags().StringVar(&crdUserSpec.CommonName, "commonName", "", "")
	createUserCmd.PersistentFlags().StringVar(&crdUserSpec.Email, "email", "", "")
	createUserCmd.PersistentFlags().StringVar(&crdUserSpec.Comment, "comment", "", "")
	createUserCmd.PersistentFlags().StringVar(&crdUserSpec.PasswordHash, "passwordHash", "", "")
	createUserCmd.PersistentFlags().IntVar(&uid, "uid", 0, "")
	if createUserCmd.PersistentFlags().Lookup("uid").Changed {
		crdUserSpec.Uid = &uid
	}
}


var createUserCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{},
	Short:   "Create new user (Admin)",
	Hidden:  false,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("ERROR: A username must be provided!\n")
			os.Exit(2)
		}
		initHttpConnection()
		userName := args[0]
		token := retrieveToken()
		if token == "" {
			token = doLogin("", "")
		}
		body, err := json.Marshal(crdUserSpec)
		response, err := httpConnection.Do("POST", fmt.Sprintf("/auth/v1/admin/%s/users/%s", provider, userName) , &internal.HttpAuth{Token: token},  bytes.NewBuffer(body))
		if err != nil {
			panic(err)
		}
		if response.StatusCode == http.StatusCreated {
			fmt.Printf("User created sucessfully.\n")
		} else {
			printHttpResponseMessage(response)
		}
		if response.StatusCode != http.StatusCreated {
			os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
		}
	},
}

