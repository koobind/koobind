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
package token

import (
	"fmt"
	"github.com/koobind/koobind/koocli/cmd/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

func init() {
	tokenDeleteCmd.SetUsageTemplate(tokenDeleteCmd.UsageTemplate())
}

var tokenDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a token (Unlog the user)",
	Run: func(cmd *cobra.Command, args []string) {
		common.InitHttpConnection()
		if len(args) != 1 {
			fmt.Printf("ERROR: A tokenBag must be provided!\n")
			os.Exit(2)
		}
		targetToken := args[0]
		tokenBag := common.RetrieveTokenBag()
		if tokenBag == nil {
			tokenBag = common.DoLogin("", "")
		}
		response, err := common.HttpConnection.Do("DELETE", "/auth/v1/admin/tokens/"+targetToken, &internal.HttpAuth{Token: tokenBag.Token}, nil)
		if err != nil {
			panic(err)
		}
		if response.StatusCode == http.StatusOK {
			fmt.Printf("Token %s is successfully deleted\n", targetToken)
		} else if response.StatusCode == http.StatusNotFound {
			fmt.Printf("ERROR: Token %s does not exists\n", targetToken)
		} else {
			common.PrintHttpResponseMessage(response)
		}
		if response.StatusCode != http.StatusOK {
			os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
		}
	},
}
