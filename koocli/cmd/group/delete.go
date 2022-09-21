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
package group

import (
	"fmt"
	. "github.com/koobind/koobind/koocli/cmd/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

func init() {
	groupDeleteCmd.PersistentFlags().StringVar(&Provider, "provider", "_", "")
}

// kubectl koo delete group grp1

var groupDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{},
	Short:   "Delete group (Admin)",
	Hidden:  false,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("ERROR: A group name must be provided!\n")
			os.Exit(2)
		}
		InitHttpConnection()
		groupName := args[0]
		token := RetrieveToken()
		if token == "" {
			token = DoLogin("", "")
		}
		response, err := HttpConnection.Do("DELETE", fmt.Sprintf("/auth/v1/admin/%s/groups/%s", Provider, groupName), &internal.HttpAuth{Token: token}, nil)
		if err != nil {
			panic(err)
		}
		if response.StatusCode == http.StatusOK {
			fmt.Printf("Group deleted successfully.\n")
		} else {
			PrintHttpResponseMessage(response)
		}
		if response.StatusCode != http.StatusOK {
			os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
		}
	},
}
