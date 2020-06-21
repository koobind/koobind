package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/koobind/koobind/koocli/cmd/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/koobind/koobind/koomgr/apis/directory/v1alpha1"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var crdUserSpec *v1alpha1.UserSpec
var uid int



func initUserParams(cmd *cobra.Command) {
	crdUserSpec = &v1alpha1.UserSpec{
	}
	cmd.PersistentFlags().StringVar(&Provider, "provider", "_", "")
	cmd.PersistentFlags().BoolVar(&crdUserSpec.Disabled, "disabled", false, "")
	cmd.PersistentFlags().StringVar(&crdUserSpec.CommonName, "commonName", "", "")
	cmd.PersistentFlags().StringVar(&crdUserSpec.Email, "email", "", "")
	cmd.PersistentFlags().StringVar(&crdUserSpec.Comment, "comment", "", "")
	cmd.PersistentFlags().StringVar(&crdUserSpec.PasswordHash, "passwordHash", "", "")
	cmd.PersistentFlags().IntVar(&uid, "uid", 0, "")
}

func applyUserCommand(cmd *cobra.Command, args []string, method string) {
	if len(args) != 1 {
		fmt.Printf("ERROR: A username must be provided!\n")
		os.Exit(2)
	}
	InitHttpConnection()
	userName := args[0]
	token := RetrieveToken()
	if token == "" {
		token = DoLogin("", "")
	}
	if cmd.PersistentFlags().Lookup("uid").Changed {
		fmt.Printf("Set uid to %d\n", uid)
		crdUserSpec.Uid = &uid
	}
	body, err := json.Marshal(crdUserSpec)
	response, err := HttpConnection.Do(method, fmt.Sprintf("/auth/v1/admin/%s/users/%s", Provider, userName) , &internal.HttpAuth{Token: token},  bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	if response.StatusCode == http.StatusCreated {
		fmt.Printf("User created sucessfully.\n")
	} else if response.StatusCode == http.StatusOK {
		fmt.Printf("User updated sucessfully.\n")
	} else {
		PrintHttpResponseMessage(response)
	}
	if response.StatusCode != http.StatusCreated {
		os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
	}

}
