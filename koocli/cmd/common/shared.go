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
package common

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
)


// Global variables. Shared by all commands
var HttpConnection *internal.HttpConnection
// package logger
var Log *logrus.Entry

var Context string
var Config *internal.Config

// Variable shared by at least two packages
var JsonOutput bool
var Provider string



func InitHttpConnection() {
	HttpConnection = internal.NewHttpConnection(Config.Server, Config.RootCaFile, Log)
}


func DoLogin(login, password string) (token string) {
	var getTokenResponse *GetTokenResponse
	for i := 0; i < 3; i++ {
		login, password = inputCredentials(login, password)
		getTokenResponse = getTokenFor(login, password)
		if getTokenResponse != nil {
			_, _ = fmt.Fprintf(os.Stdout, "logged successfully..\n")
			internal.SaveTokenBag(Context, &internal.TokenBag{
				Token:      getTokenResponse.Token,
				ClientTTL:  getTokenResponse.ClientTTL,
				LastAccess: time.Now(),
			} )
			Log.Debugf("TokenResponse:%v\n", getTokenResponse)
			return getTokenResponse.Token
		}
		_, _ = fmt.Fprintf(os.Stderr, "Invalid login!\n")
		login = ""; password = ""
	}
	_, _ = fmt.Fprintf(os.Stderr, "Too many failure !!!\n")
	return ""
}

func getTokenFor(login, password string) *GetTokenResponse {
	response, err := HttpConnection.Do("GET", "/auth/v1/getToken", &internal.HttpAuth{Login: login, Password: password}, nil)
	if err != nil {
		panic(err)
	}
	if response.StatusCode == http.StatusOK {
		var getTokenResponse GetTokenResponse
		err = json.NewDecoder(response.Body).Decode(&getTokenResponse)
		if err != nil {
			panic(err)
		}
		return &getTokenResponse
	} else if response.StatusCode == http.StatusUnauthorized {
		return nil
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Invalid http response: %s, (Status:%d)\n", response.Status, response.StatusCode)
		return nil
		//panic(fmt.Errorf("Invalid http response: %s, (Status:%d)\n", response.Status, response.StatusCode))
	}
}

func inputCredentials(login, password string) (string, string) {
	if login == "" {
		_, err := fmt.Fprint(os.Stderr, "Login:")
		if err != nil {
			panic(err)
		}
		r := bufio.NewReader(os.Stdin)
		login, err = r.ReadString('\n')
		if err != nil {
			panic(err)
		}
		login = strings.TrimSpace(login)
	}
	if password == "" {
		_, err := fmt.Fprint(os.Stderr, "Password:")
		if err != nil {
			panic(err)
		}
		bytePassword, err2 := terminal.ReadPassword(int(syscall.Stdin))
		if err2 != nil {
			panic(err2)
		}
		_, _ = fmt.Fprintf(os.Stderr, "\n")
		password = strings.TrimSpace(string(bytePassword))
	}
	return login, password
}

func ValidateToken(token string) *User {
	validateTokenRequest := ValidateTokenRequest{
		ApiVersion: "",
		Kind:       "",
	}
	validateTokenRequest.Spec.Token = token
	body, err := json.Marshal(validateTokenRequest)
	if err != nil {
		panic(err)
	}
	// Will use the service intended for the authentication webhook, but with a GET method
	response, err2 := HttpConnection.Do("GET", "/auth/v1/validateToken", nil, bytes.NewBuffer(body))
	if err2 != nil {
		panic(err2)
	}
	if response.StatusCode == http.StatusOK {
		var validateTokenResponse ValidateTokenResponse
		err = json.NewDecoder(response.Body).Decode(&validateTokenResponse)
		if err != nil {
			panic(err)
		}
		if validateTokenResponse.Status.Authenticated {
			return validateTokenResponse.Status.User
		} else {
			return nil
		}
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Invalid http response: %s, (Status:%d)\n", response.Status, response.StatusCode)
		return nil
		//panic(fmt.Errorf("Invalid http response: %s, (Status:%d)\n", response.Status, response.StatusCode))
	}
}


// Retrieve the token locally, or, if expired, validate again against the server. Return "" if there is no valid token
func RetrieveToken() string {
	tokenBag := internal.LoadTokenBag(Context)
	if tokenBag != nil {
		now := time.Now()
		if now.Before(tokenBag.LastAccess.Add(tokenBag.ClientTTL.Duration)) {
			// tokenBag still valid
			return tokenBag.Token
		} else {
			if user := ValidateToken(tokenBag.Token); user != nil {
				tokenBag.LastAccess = time.Now()
				internal.SaveTokenBag(Context, tokenBag)
				return tokenBag.Token
			} else {
				internal.DeleteTokenBag(Context)
				return ""
			}
		}
	} else {
		return ""
	}
}

func PrintHttpResponseMessage(response *http.Response) {
	data, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode == http.StatusForbidden {
		fmt.Printf("ERROR: You are not allowed to perform this operation!\n")
	} else if response.StatusCode == http.StatusUnauthorized {
		fmt.Printf("ERROR: Unable to authenticate!\n")
	} else {
		if data != nil && len(data) > 0 {
			m := strings.TrimSpace(string(data))
			fmt.Printf("ERROR: %s: %s.", response.Status, m)
		} else {
			fmt.Printf("ERROR: %s.", response.Status)
		}
		if response.StatusCode > http.StatusInternalServerError {
			fmt.Printf(" Check server logs or contact your server administrator.")
		} else {
			fmt.Print("\n")
		}
	}
}

