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
	"fmt"
	"github.com/koobind/koobind/koocli/cmd/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/koobind/koobind/koomgr/apis/proto"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
)

var explainAuth bool

func init() {
	DescribeUserCmd.PersistentFlags().BoolVarP(&common.JsonOutput, "json", "", false, "Output in JSON")
	DescribeUserCmd.PersistentFlags().BoolVar(&explainAuth, "explain", false, "Explain user authentication")
}

var DescribeUserCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"users"},
	Short:   "Describe a specified user (admin)",
	Hidden:  false,
	//Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("ERROR: A username must be provided!\n")
			os.Exit(2)
		}
		common.InitHttpConnection()
		userName := args[0]
		token := common.RetrieveToken()
		if token == "" {
			token = common.DoLogin("", "")
		}
		response, err := common.HttpConnection.Do("GET", "/auth/v1/admin/users/"+userName, &internal.HttpAuth{Token: token}, nil)
		if err != nil {
			panic(err)
		}
		if response.StatusCode == http.StatusOK {
			if common.JsonOutput {
				data, _ := ioutil.ReadAll(response.Body)
				fmt.Print(string(data))
			} else {
				var userDescribeResponse proto.UserDescribeResponse
				err = json.NewDecoder(response.Body).Decode(&userDescribeResponse)
				if err != nil {
					panic(err)
				}
				tw := new(tabwriter.Writer)
				tw.Init(os.Stdout, 2, 4, 3, ' ', 0)
				if explainAuth {
					_, _ = fmt.Fprintf(tw, "PROVIDER\tFOUND\tAUTH\tUID\tGROUPS\tEMAIL\tCOMMON NAME\tCOMMENT")
					//authorityFound := false
					for _, userEntry := range userDescribeResponse.User.Entries {
						var found = ""
						var authority = ""
						if userEntry.Found {
							found = "*"
							if userEntry.Authority {
								if userEntry.ProviderName == userDescribeResponse.User.Authority {
									authority = "*"
								} else {
									authority = "+"
								}
							}
						}
						_, _ = fmt.Fprintf(tw, "\n%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", userEntry.ProviderName, found, authority, userEntry.Uid, array2String(userEntry.Groups), userEntry.Email, userEntry.CommonName, array2String(userEntry.Messages))
					}
				} else {
					_, _ = fmt.Fprintf(tw, "USER\tID\tGROUPS\tAUTHORITY")
					_, _ = fmt.Fprintf(tw, "\n%s\t%s\t%s\t%s", userDescribeResponse.User.Name, userDescribeResponse.User.Uid, strings.Join(userDescribeResponse.User.Groups, ","), userDescribeResponse.User.Authority)
				}
				_, _ = fmt.Fprintf(tw, "\n")
				_ = tw.Flush()
			}
		} else if response.StatusCode == http.StatusNotFound {
			fmt.Printf("ERROR: User %s does not exists!\n", userName)
		} else {
			common.PrintHttpResponseMessage(response)
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
