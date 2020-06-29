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
package group

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

var (
	crdGroupSpec =  &v1alpha1.GroupSpec{}
	disabled     bool
	enabled      bool
	_true        = true
	_false       = false
)

func initGroupParams(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&Provider, "provider", "_", "")
	cmd.PersistentFlags().BoolVar(&disabled, "disabled", false, "")
	cmd.PersistentFlags().BoolVar(&enabled, "enabled", false, "")
	cmd.PersistentFlags().StringVar(&crdGroupSpec.Description, "description", "", "")
}

func applyGroupCommand(cmd *cobra.Command, args []string, method string) {
	if len(args) != 1 {
		fmt.Printf("ERROR: A group name must be provided!\n")
		os.Exit(2)
	}
	if enabled && disabled {
		fmt.Printf("ERROR: A group can be both enabled and disabled!\n")
		os.Exit(2)
	}
	InitHttpConnection()
	groupName := args[0]
	token := RetrieveToken()
	if token == "" {
		token = DoLogin("", "")
	}
	if enabled {
		crdGroupSpec.Disabled = &_false
	}
	if disabled {
		crdGroupSpec.Disabled = &_true
	}
	body, err := json.Marshal(crdGroupSpec)
	response, err := HttpConnection.Do(method, fmt.Sprintf("/auth/v1/admin/%s/groups/%s", Provider, groupName) , &internal.HttpAuth{Token: token},  bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	if response.StatusCode == http.StatusCreated {
		fmt.Printf("Group created successfully.\n")
	} else if response.StatusCode == http.StatusOK {
		fmt.Printf("Group updated successfully.\n")
	} else {
		PrintHttpResponseMessage(response)
	}
	if response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusOK {
		os.Exit(internal.ReturnCodeFromStatusCode(response.StatusCode))
	}
}
