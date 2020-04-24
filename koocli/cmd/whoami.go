package cmd

import (
	"fmt"
	. "github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var displayToken bool

func init() {
	rootCmd.AddCommand(whoamiCmd)
	whoamiCmd.PersistentFlags().BoolVarP(&displayToken, "token", "", false, "Display token")
}

var whoamiCmd = &cobra.Command{
	Use:	"whoami",
	Short:  "Display current logged user, if any",
	Run:    func(cmd *cobra.Command, args []string) {
		initHttpConnection()
		tokenBag := internal.LoadTokenBag(context)
		if user := getUser(tokenBag); user != nil {
			if displayToken {
				fmt.Printf("user:%s  id:%s  groups:%s  token:%s\n", user.Username, user.Uid, strings.Join(user.Groups, ","), tokenBag.Token)
			} else {
				fmt.Printf("user:%s  id:%s  groups:%s\n", user.Username, user.Uid, strings.Join(user.Groups, ","))
			}
		} else {
			fmt.Printf("Nobody! (Not logged)\n")
			os.Exit(3)
		}
	},
}


// getUser() trigger a server exchange (validateToken) in all cases, as we have no local storage of user info.
func getUser(tokenBag *internal.TokenBag) *User {
	if tokenBag != nil {
		user := validateToken(tokenBag.Token)
		if user != nil {
			return user
		}
	}
	return nil
}

