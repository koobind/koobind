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
			fmt.Printf("ERROR: Invalid http response: %d, (Status:%d) Contact server administrator\n", response.Status, response.StatusCode)
		}
		if response.StatusCode != http.StatusOK {
			os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
		}
	},
}

