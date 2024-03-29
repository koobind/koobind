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
package groupbinding

import (
	"fmt"
	. "github.com/koobind/koobind/koocli/cmd/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

func init() {
	groupBindingDeleteCmd.PersistentFlags().StringVar(&Provider, "provider", "_", "")
}

// kubectl koo delete group grp1

var groupBindingDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{},
	Short:   "Delete groupBinding (Admin)",
	Hidden:  false,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Printf("ERROR: A user name and a group name must be provided!\n")
			os.Exit(2)
		}
		InitHttpConnection()
		userName := args[0]
		groupName := args[1]
		tokenBag := RetrieveTokenBag()
		if tokenBag == nil {
			tokenBag = DoLogin("", "")
		}
		response, err := HttpConnection.Do("DELETE", fmt.Sprintf("/auth/v1/admin/%s/groupbindings/%s/%s", Provider, userName, groupName), &internal.HttpAuth{Token: tokenBag.Token}, nil)
		if err != nil {
			panic(err)
		}
		if response.StatusCode == http.StatusOK {
			fmt.Printf("GroupBinding deleted successfully.\n")
		} else {
			PrintHttpResponseMessage(response)
		}
		if response.StatusCode != http.StatusOK {
			os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
		}
	},
}
