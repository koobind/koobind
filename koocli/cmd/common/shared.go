/*
Copyright (C) 2020 Serge ALEXANDRE

# This file is part of koobind project

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
	"github.com/koobind/koobind/koocli/internal"
	proto_v2 "github.com/koobind/koobind/koomgr/apis/proto/auth/v2"
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

func DoLogin(login, password string) *internal.TokenBag {
	tokenBag := DoLoginSilently(login, password)
	if tokenBag != nil {
		_, _ = fmt.Fprintf(os.Stdout, "logged successfully..\n")
	}
	return tokenBag
}

// Warning: As this function is used by the 'auth' command, which send json result to stdout, it may only send prompt to stderr
func DoLoginSilently(login, password string) *internal.TokenBag {
	maxTry := 3
	if login != "" && password != "" {
		maxTry = 1 // If all is provided on command line, do not prompt in case of failure
	}
	for i := 0; i < maxTry; i++ {
		login, password = inputCredentials(login, password)
		loginResponse := loginAndGetToken(login, password)
		if loginResponse != nil {
			tokenBag := &internal.TokenBag{
				Token:      loginResponse.Token,
				ClientTTL:  loginResponse.ClientTTL,
				LastAccess: time.Now(),
				Username:   loginResponse.Username,
				Uid:        loginResponse.Uid,
				Groups:     loginResponse.Groups,
			}
			internal.SaveTokenBag(Context, tokenBag)
			Log.Debugf("TokenResponse:%v\n", loginResponse)
			return tokenBag
		}
		_, _ = fmt.Fprintf(os.Stderr, "Invalid login!\n")
		login = ""
		password = ""
	}
	if maxTry > 1 {
		_, _ = fmt.Fprintf(os.Stderr, "Too many failure !!!\n")
	}
	return nil
}

func loginAndGetToken(login, password string) *proto_v2.LoginResponse {
	loginRequestPayload := proto_v2.LoginRequest{
		Login:         login,
		Password:      password,
		GenerateToken: true,
		Client:        Config.Client,
	}
	body, err := json.Marshal(loginRequestPayload)
	if err != nil {
		panic(err)
	}
	response, err := HttpConnection.Do("POST", proto_v2.LoginUrlPath, nil, bytes.NewBuffer(body))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to exchnage with authentication server: %v. Check server logs\n", err)
		return nil
		//os.Exit(2)
	}
	if response.StatusCode == http.StatusUnauthorized {
		return nil
	}
	if response.StatusCode != http.StatusOK {
		var b bytes.Buffer
		b.ReadFrom(response.Body)
		_, _ = fmt.Fprintf(os.Stderr, "Invalid http response: %s, (Status:%d) %v\n", response.Status, response.StatusCode, b.String())
		return nil
		//panic(fmt.Errorf("Invalid http response: %s, (Status:%d)\n", response.Status, response.StatusCode))

	}
	var loginResponse proto_v2.LoginResponse
	err = json.NewDecoder(response.Body).Decode(&loginResponse)
	if err != nil {
		var b bytes.Buffer
		b.ReadFrom(response.Body)
		_, _ = fmt.Fprintf(os.Stderr, "Unable to decode server response '%s': %v\n", b.String(), err)
		panic(err)
	}
	return &loginResponse
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
		password = inputPassword("Password:")
	}
	return login, password
}

func inputPassword(prompt string) string {
	_, err := fmt.Fprint(os.Stderr, prompt)
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

func ValidateToken(token string) bool {
	validateTokenRequest := proto_v2.ValidateTokenRequest{
		Token:  token,
		Client: Config.Client,
	}
	body, err := json.Marshal(validateTokenRequest)
	if err != nil {
		panic(err)
	}
	response, err2 := HttpConnection.Do("POST", proto_v2.ValidateTokenUrlPath, nil, bytes.NewBuffer(body))
	if err2 != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to connect to authentication server: %v\n", err2)
		os.Exit(2)
	}
	if response.StatusCode == http.StatusOK {
		var validateTokenResponse proto_v2.ValidateTokenResponse
		err = json.NewDecoder(response.Body).Decode(&validateTokenResponse)
		if err != nil {
			panic(err)
		}
		return validateTokenResponse.Valid
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Invalid http response: %s, (Status:%d)\n", response.Status, response.StatusCode)
		return false
		//panic(fmt.Errorf("Invalid http response: %s, (Status:%d)\n", response.Status, response.StatusCode))
	}
}

// Retrieve the token locally, or, if expired, validate again against the server. Return "" if there is no valid token
func RetrieveTokenBag() *internal.TokenBag {
	tokenBag := internal.LoadTokenBag(Context)
	if tokenBag != nil {
		now := time.Now()
		if now.Before(tokenBag.LastAccess.Add(tokenBag.ClientTTL)) {
			// tokenBag still valid
			return tokenBag
		} else {
			if ValidateToken(tokenBag.Token) {
				tokenBag.LastAccess = time.Now()
				internal.SaveTokenBag(Context, tokenBag)
				return tokenBag
			} else {
				internal.DeleteTokenBag(Context)
				return nil
			}
		}
	} else {
		return nil
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
			fmt.Printf("ERROR: %s: %s", response.Status, m)
		} else {
			fmt.Printf("ERROR: %s", response.Status)
		}
		if response.StatusCode > http.StatusInternalServerError {
			fmt.Printf(" Check server logs or contact your server administrator.")
		} else {
			fmt.Print("\n")
		}
	}
}
