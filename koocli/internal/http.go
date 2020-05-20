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
package internal

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Concurrent safe, as http.Client is and baseUrl is not mutated.
type HttpConnection struct {
	httpClient *http.Client
	baseUrl string
}


type HttpAuth struct {
	Login string
	Password string
	Token string
}


func NewHttpConnection(baseUrl string, rootCaFile string) *HttpConnection {
	baseUrl = strings.TrimRight(baseUrl, "/")		// No trailing '/'
	if rootCaFile == "" {
		return &HttpConnection{
			httpClient: http.DefaultClient,
			baseUrl: baseUrl,
		}
	} else {
		rootPEM, err := ioutil.ReadFile(rootCaFile)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Unable to read %s: %s\n\n", rootCaFile, err)
			os.Exit(2)
		}
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(rootPEM))
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Unable to parse CA certificate (%s)\n\n", rootCaFile)
			os.Exit(2)
		}
		tlsConf := &tls.Config{RootCAs: roots}
		tr := &http.Transport{TLSClientConfig: tlsConf}
		return &HttpConnection {
			httpClient: &http.Client{Transport: tr},
			baseUrl: baseUrl,
		}
	}
}

func (this *HttpConnection) Get(urlPath string, auth *HttpAuth, body io.Reader) (*http.Response, error) {
	targetUrl := this.baseUrl + urlPath
	//fmt.Printf("baseUrl:'%s'   urlPath:'%s'   targetUrl:'%s'\n", this.baseUrl, urlPath, targetUrl)
	request, err := http.NewRequest("GET", targetUrl, body)
	if err != nil {
		return nil, err
	}
	if auth != nil {
		if auth.Login != "" {
			request.SetBasicAuth(auth.Login, auth.Password)
		}
		if auth.Token != "" {
			request.Header.Set("Authorization", "Bearer "+ auth.Token)
		}
	}
	return this.httpClient.Do(request)
}

func (this *HttpConnection) Delete(urlPath string, auth *HttpAuth) (*http.Response, error) {
	targetUrl := this.baseUrl + urlPath
	request, err := http.NewRequest("DELETE", targetUrl, nil)
	if err != nil {
		return nil, err
	}
	if auth != nil {
		if auth.Login != "" {
			request.SetBasicAuth(auth.Login, auth.Password)
		}
		if auth.Token != "" {
			request.Header.Set("Authorization", "Bearer "+ auth.Token)
		}
	}
	return this.httpClient.Do(request)
}



func ReturnCodeFromStatusCode(sc int) int {
	if sc > 400 && sc < 425 {
		return sc - 400
	} else {
		return 125
	}
}

