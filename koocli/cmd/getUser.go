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
	"encoding/json"
	"fmt"
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
)

var explainAuth bool

func init() {
	getCmd.AddCommand(usersCmd)
	getCmd.PersistentFlags().BoolVarP(&explainAuth, "explain", "", false, "Explain user authentication")
}

var usersCmd = &cobra.Command{
	Use:	"user",
	Aliases: []string{"users"},
	Short:  "Describe a specified user (admin)",
	Hidden: false,
	//Args: cobra.MinimumNArgs(1),
	Run:    func(cmd *cobra.Command, args []string) {
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
		response, err := httpConnection.Get(common.V1Admin + "users/" + userName, &internal.HttpAuth{Token: token},  nil)
		if err != nil {
			panic(err)
		}
		if response.StatusCode == http.StatusOK {
			if jsonOutput {
				data, _ := ioutil.ReadAll(response.Body)
				fmt.Print(string(data))
			} else {
				var userDescribeResponse common.UserDescribeResponse
				err = json.NewDecoder(response.Body).Decode(&userDescribeResponse)
				if err != nil {
					panic(err)
				}
				tw := new(tabwriter.Writer)
				tw.Init(os.Stdout, 2, 4, 3, ' ', 0)
				if explainAuth {
					_, _ = fmt.Fprintf(tw, "PROVIDER\tFOUND\tAUTH\tUID\tGROUPS\tEMAIL\tCOMMON NAME\tCOMMENT")
					//authorityFound := false
					for _, userStatus := range userDescribeResponse.UserStatuses {
						var found = ""
						var authority = ""
						if userStatus.Found {
							found = "*"
							if userStatus.Authority {
								if userStatus.ProviderName == userDescribeResponse.Authority {
									authority = "*"
								} else {
									authority = "+"
								}
							}
						}
						_, _ = fmt.Fprintf(tw, "\n%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", userStatus.ProviderName, found, authority, userStatus.Uid, array2String(userStatus.Groups), userStatus.Email, userStatus.CommonName, array2String(userStatus.Messages))
					}
				} else {
					_, _ = fmt.Fprintf(tw, "USER\tID\tGROUPS\tAUTHORITY")
					_, _ = fmt.Fprintf(tw, "\n%s\t%s\t%s\t%s", userDescribeResponse.User.Username, userDescribeResponse.User.Uid, strings.Join(userDescribeResponse.User.Groups, ","), userDescribeResponse.Authority)
				}
				_, _ = fmt.Fprintf(tw, "\n")
				_ = tw.Flush()
			}
		} else if response.StatusCode == http.StatusForbidden {
			fmt.Printf("ERROR: You are not allowed to perform this operation!\n")
		} else if response.StatusCode == http.StatusUnauthorized {
			fmt.Printf("ERROR: Unable to authenticate!\n")
		} else {
			fmt.Printf("ERROR: Invalid http response: %d, (Status:%d) Contact server administrator\n", response.Status, response.StatusCode)
		}
		if response.StatusCode != http.StatusOK {
			os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
		}
	},
}

func array2String(a []string) string {
	if a == nil || len(a) == 0 {
		return ""
	}
	return fmt.Sprintf("[%s]", strings.Join(a, ","))
}
