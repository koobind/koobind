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
package token

import (
	"encoding/json"
	"fmt"
	"github.com/koobind/koobind/common"
	. "github.com/koobind/koobind/koocli/cmd/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
)

func init() {
	GetTokensCmd.PersistentFlags().BoolVarP(&JsonOutput, "json", "", false, "Output in JSON")
}


var GetTokensCmd = &cobra.Command{
	Use:	"token",
	Aliases: []string{"tokens"},
	Short:  "List currently active token (Admin)",
	Hidden: false,
	Run:    func(cmd *cobra.Command, args []string) {
		InitHttpConnection()
		token := RetrieveToken()
		if token == "" {
			token = DoLogin("", "")
		}
		response, err := HttpConnection.Do("GET", "/auth/v1/admin/tokens", &internal.HttpAuth{Token: token},  nil)
		if err != nil {
			panic(err)
		}
		if response.StatusCode == http.StatusOK {
			if JsonOutput {
				data, _ := ioutil.ReadAll(response.Body)
				fmt.Print(string(data))
			} else {
				var tokenListResponse common.TokenListResponse
				err = json.NewDecoder(response.Body).Decode(&tokenListResponse)
				if err != nil {
					panic(err)
				}
				//fmt.Print(tokenListResponse)
				tw := new(tabwriter.Writer)
				tw.Init(os.Stdout, 2, 4, 3, ' ', 0)
				_, _ = fmt.Fprintf(tw, "TOKEN\tUSER\tUID\tGROUPS\tCREATED ON\tLAST HIT")
				for _, ut := range tokenListResponse.Tokens {
					_, _ = fmt.Fprintf(tw, "\n%s\t%s\t%s\t%s\t%s\t%s", ut.Token, ut.User.Username, ut.User.Uid, strings.Join(ut.User.Groups, ","), ut.Creation.Format("01-02 15:04:05"), ut.LastHit.Format("15:04:05"))
				}
				_, _ = fmt.Fprintf(tw, "\n")
				_ = tw.Flush()
			}
		} else {
			PrintHttpResponseMessage(response)
		}
		if response.StatusCode != http.StatusOK {
			os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
		}
	},
}

