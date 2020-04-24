package cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(authCmd)
}

var authCmd = &cobra.Command{
	Use:	"auth",
	Short:  "To be used as client-go exec plugin",
	Hidden: true,
	Run:    func(cmd *cobra.Command, args []string) {
		initHttpConnection()
		token := retrieveToken()
		if token == "" {
			token = doLogin("", "")
		}
		ec := ExecCredential{
			ApiVersion: "client.authentication.k8s.io/v1beta1",
			Kind: "ExecCredential",
		}
		if token == "" {
			// No token
		} else {
			ec.Status.Token = token
		}
		err := json.NewEncoder(os.Stdout).Encode(ec)
		if err != nil {
			panic(err)
		}
	},
}



type ExecCredential struct {
	ApiVersion string	`json:"apiVersion"`
	Kind string			`json:"kind"`
	Status struct {
		Token string	`json:"token"`
	}					`json:"status"`
}



