package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"
)
// https://medium.com/@jcox250/password-hash-salt-using-golang-b041dc94cb72

var hash_password string

func init() {
	rootCmd.AddCommand(hashCmd)
	hashCmd.PersistentFlags().StringVarP(&hash_password, "password", "", "", "User password")

}

var hashCmd = &cobra.Command{
	Use:	"hash",
	Short:  "Provided password hash, for use in config file",
	Run:    func(cmd *cobra.Command, args []string) {
		if hash_password == "" {
			hash_password = inputPassword( "Password:")
			password2 := inputPassword( "Confirm password:")
			if hash_password != password2 {
				fmt.Printf("Passwords did not match!\n")
				return
			}
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(hash_password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", string(hash))
	},
}


func inputPassword(prompt string) string {
	_, err := fmt.Fprint(os.Stdout, prompt)
	if err != nil {
		panic(err)
	}
	bytePassword, err2 := terminal.ReadPassword(int(syscall.Stdin))
	if err2 != nil {
		panic(err2)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\n")
	return strings.TrimSpace(string(bytePassword))
}




