package cmd

import (
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"os"
)


var login string
var login_password string

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.PersistentFlags().StringVarP(&login, "user", "", "", "User name")
	loginCmd.PersistentFlags().StringVarP(&login_password, "password", "", "", "User password")

}


var loginCmd = &cobra.Command{
	Use:	"login",
	Short:  "Logout and get a new token",
	Run:    func(cmd *cobra.Command, args []string) {
		initHttpConnection()
		internal.DeleteTokenBag(context)	// Logout first. Don't stay logged with old token if we are unable to login
		t := doLogin(login, login_password)
		if t == "" {
			os.Exit(3)
		}
	},
}
