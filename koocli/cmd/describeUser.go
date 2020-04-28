package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"text/tabwriter"
)

func init() {
	describeCmd.AddCommand(usersCmd)
}

var usersCmd = &cobra.Command{
	Use:	"user",
	//Aliases: []string{"users"},
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
			var userDescribeResponse common.UserDescribeResponse
			err = json.NewDecoder(response.Body).Decode(&userDescribeResponse)
			if err != nil {
				panic(err)
			}
			tw := new(tabwriter.Writer)
			tw.Init(os.Stdout, 2, 4, 1, ' ', 0)
			_, _ = fmt.Fprintf(tw, "PROVIDER\tFOUND\tAUTH\tUID\tGROUPS")
			for _, userStatus := range userDescribeResponse.UserStatuses {
				var found = ""
				if userStatus.Found { found = "*" }
				var password = ""
				if userStatus.PasswordStatus == common.Wrong { password = "*" }
				_, _ = fmt.Fprintf(tw, "\n%s\t%s\t%s\t%s\t%v", userStatus.ProviderName, found, password, userStatus.Uid, userStatus.Groups)
			}
			_, _ = fmt.Fprintf(tw, "\n")
			_ = tw.Flush()
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

