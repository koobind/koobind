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
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)


func init() {
	cancelCmd.AddCommand(cancelTokenCmd)
}

var cancelTokenCmd = &cobra.Command{
	Use:	"token <token>",
	Short:  "Cancel a token (Unlog the user)",
	Args: cobra.MinimumNArgs(1),
	Run:    func(cmd *cobra.Command, args []string) {
		initHttpConnection()
		targetToken := args[0]
		token := retrieveToken()
		if token == "" {
			token = doLogin("", "")
		}
		response, err := httpConnection.Delete(common.V1Admin + "tokens/" + targetToken, &internal.HttpAuth{Token: token})
		if err != nil {
			panic(err)
		}
		if response.StatusCode == http.StatusOK {
			fmt.Printf("Token %s is successfully cancelled\n", targetToken)
		} else if response.StatusCode == http.StatusNotFound {
			fmt.Printf("ERROR: Token %s does not exists\n", targetToken)
		} else if response.StatusCode == http.StatusForbidden {
			fmt.Printf("ERROR: You are not allowed to perform this operation!\n")
		} else if response.StatusCode == http.StatusUnauthorized {
			fmt.Printf("ERROR: Unable to authenticate!\n")
		} else {
			fmt.Printf("ERROR: Invalid http response: %s, (Status:%d) Contact server administrator\n", response.Status, response.StatusCode)
		}
		if response.StatusCode != http.StatusOK {
			os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
		}
	},
}

