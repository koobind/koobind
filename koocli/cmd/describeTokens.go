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

func init() {
	describeCmd.AddCommand(tokensCmd)
}

var tokensCmd = &cobra.Command{
	Use:	"tokens",
	//Aliases: []string{"tokens"},
	Short:  "List currently active token (Admin)",
	Hidden: false,
	Run:    func(cmd *cobra.Command, args []string) {
		initHttpConnection()
		token := retrieveToken()
		if token == "" {
			token = doLogin("", "")
		}
		response, err := httpConnection.Get(common.V1Admin + "tokens", &internal.HttpAuth{Token: token},  nil)
		if err != nil {
			panic(err)
		}
		if response.StatusCode == http.StatusOK {
			if jsonOutput {
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
				tw.Init(os.Stdout, 2, 4, 1, ' ', 0)
				_, _ = fmt.Fprintf(tw, "TOKEN\tUSER\tUID\tGROUPS\tCREATED ON\tLAST HIT")
				for _, ut := range tokenListResponse.Tokens {
					_, _ = fmt.Fprintf(tw, "\n%s\t%s\t%s\t%s\t%s\t%s", ut.Token, ut.User.Username, ut.User.Uid, strings.Join(ut.User.Groups, ","), ut.Creation.Format("01-02 15:04:05"), ut.LastHit.Format("15:04:05"))
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

