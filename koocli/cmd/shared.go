package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koocli/internal"
	"golang.org/x/crypto/ssh/terminal"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
)

var httpConnection *internal.HttpConnection

func initHttpConnection() {
	httpConnection = internal.NewHttpConnection(config.Server, config.RootCaFile)
}


func doLogin(login, password string) (token string) {
	var getTokenResponse *GetTokenResponse
	for i := 0; i < 3; i++ {
		login, password = inputCredentials(login, password)
		getTokenResponse = getTokenFor(login, password)
		if getTokenResponse != nil {
			_, _ = fmt.Fprintf(os.Stderr, "logged successfully..\n")
			internal.SaveTokenBag(context, &internal.TokenBag{
				Token:      getTokenResponse.Token,
				ClientTTL:  getTokenResponse.ClientTTL,
				LastAccess: time.Now(),
			} )
			log.Debugf("TokenResponse:%v\n", getTokenResponse)
			return getTokenResponse.Token
		}
		_, _ = fmt.Fprintf(os.Stderr, "Invalid login!\n")
		login = ""; password = ""
	}
	_, _ = fmt.Fprintf(os.Stderr, "Too many failure !!!\n")
	return ""
}

func getTokenFor(login, password string) *GetTokenResponse {
	response, err := httpConnection.Get(V1GetToken, &internal.HttpAuth{Login: login, Password: password}, nil)
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

func validateToken(token string) *User {
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
	response, err2 := httpConnection.Get(V1ValidateTokenUrl, nil, bytes.NewBuffer(body))
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
func retrieveToken() string {
	tokenBag := internal.LoadTokenBag(context)
	if tokenBag != nil {
		now := time.Now()
		if now.Before(tokenBag.LastAccess.Add(tokenBag.ClientTTL.Duration)) {
			// tokenBag still valid
			return tokenBag.Token
		} else {
			if user := validateToken(tokenBag.Token); user != nil {
				tokenBag.LastAccess = time.Now()
				internal.SaveTokenBag(context, tokenBag)
				return tokenBag.Token
			} else {
				internal.DeleteTokenBag(context)
				return ""
			}
		}
	} else {
		return ""
	}
}

