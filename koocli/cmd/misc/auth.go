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
	"encoding/json"
	"github.com/koobind/koobind/koocli/cmd/common"
	"github.com/spf13/cobra"
	"os"
)

var AuthCmd = &cobra.Command{
	Use:    "auth",
	Short:  "To be used as client-go exec plugin",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		common.InitHttpConnection()
		token := common.RetrieveToken()
		if token == "" {
			token = common.DoLoginSilently("", "")
		}
		ec := ExecCredential{
			ApiVersion: "client.authentication.k8s.io/v1beta1",
			Kind:       "ExecCredential",
		}
		if token == "" {
			// No token
		} else {
			ec.Status.Token = token
		}
		err := json.NewEncoder(os.Stdout).Encode(ec)
		if err != nil {
			panic(err)
		}
	},
}

type ExecCredential struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Status     struct {
		Token string `json:"token"`
	} `json:"status"`
}
