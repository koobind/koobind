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

package authserver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/koobind/koobind/koomgr/internal/authserver/certwatcher"
	"github.com/koobind/koobind/koomgr/internal/authserver/handlers"
	v1 "github.com/koobind/koobind/koomgr/internal/authserver/handlers/v1"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/token"
	"net"
	"net/http"
	"os"
	"path/filepath"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
)

var serverLog = logf.Log.WithName("Auth server")

type Server struct {
	// Host is the address that the server will listen on.
	// Defaults to "" - all addresses.
	Host string

	// Port is the port number that the server will serve.
	// It will be defaulted to 443 if unspecified.
	Port int

	// CertDir is the directory that contains the server key and certificate. The
	// server key and certificate.
	CertDir string

	// CertName is the server certificate name. Defaults to tls.crt.
	CertName string

	// CertName is the server key name. Defaults to tls.key.
	KeyName string

	// WebhookMux is the multiplexer that handles different handlerByPath.
	//Mux *http.ServeMux
	Router *mux.Router

	Manager ctrlrt.Manager

	// handlerByPath keep track of all registered handlers for dependency injection,
	// and to provide better panic messages on duplicate handler registration.
	handlerByPath map[string]http.Handler
}

// setDefaults does defaulting for the Server.
func (this *Server) setDefaults() {
	this.handlerByPath = map[string]http.Handler{}
	//if this.Mux == nil {
	//	//this.Mux = http.NewServeMux()
	//	this.Mux = mux.
	//}
	if this.Router == nil {
		this.Router = mux.NewRouter()
		this.Router.Use(LogHttp)
	}

	if this.Port <= 0 {
		this.Port = 443
	}

	if len(this.CertDir) == 0 {
		this.CertDir = filepath.Join(os.TempDir(), "k8s-webhook-server", "serving-certs")
	}

	if len(this.CertName) == 0 {
		this.CertName = "tls.crt"
	}

	if len(this.KeyName) == 0 {
		this.KeyName = "tls.key"
	}
}

func (*Server) NeedLeaderElection() bool {
	return false
}

func (s *Server) Init(tokenBasket token.TokenBasket, kubeClient client.Client, providerChain providers.ProviderChain) {
	s.setDefaults()

	s.Router.Handle("/auth/v1/validateToken", &v1.ValidateTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:      ctrlrt.Log.WithName("authV1validateToken"),
			TokenBasket: tokenBasket,
		},
	}).Methods("GET", "POST") // POST is from Api server while GET is from our client

	s.Router.Handle("/auth/v1/getToken", &v1.GetTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:      ctrlrt.Log.WithName("v1getToken"),
			TokenBasket: tokenBasket,
		},
		Providers: providerChain,
	}).Methods("GET")

	s.Router.Handle("/auth/v1/changePassword", &v1.ChangePasswordHandler{
		AuthHandler: handlers.AuthHandler{
			BaseHandler: handlers.BaseHandler{
				Logger:      ctrlrt.Log.WithName("v1changePassword"),
				TokenBasket: tokenBasket,
			},
			Providers: providerChain,
		},
	}).Methods("POST")

	newAdminHandler := func(hf v1.HandlerFunc, loggerName string) *v1.AdminV1Handler {
		return &v1.AdminV1Handler{
			AuthHandler: handlers.AuthHandler{
				BaseHandler: handlers.BaseHandler{
					Logger:      ctrlrt.Log.WithName(loggerName),
					TokenBasket: tokenBasket,
				},
				Providers: providerChain,
			},
			AdminGroup:  config.Conf.AdminGroup,
			KubeClient:  kubeClient,
			HandlerFunc: hf,
		}
	}

	s.Router.Handle("/auth/v1/admin/tokens/{token}", newAdminHandler(v1.DeleteToken, "adminV1deleteToken")).Methods("DELETE")
	s.Router.Handle("/auth/v1/admin/tokens", newAdminHandler(v1.ListToken, "adminV1listToken")).Methods("GET")
	s.Router.Handle("/auth/v1/admin/users/{user}", newAdminHandler(v1.DescribeUser, "adminV1describeUser")).Methods("GET")
	s.Router.Handle("/auth/v1/admin/{provider}/users/{user}", newAdminHandler(v1.AddApplyPatchUser, "adminV1addApplyPatchUser")).Methods("POST", "PUT", "PATCH")
	s.Router.Handle("/auth/v1/admin/{provider}/users/{user}", newAdminHandler(v1.DeleteUser, "adminV1deleteUser")).Methods("DELETE")
	s.Router.Handle("/auth/v1/admin/{provider}/groups/{group}", newAdminHandler(v1.AddApplyPatchGroup, "adminV1addApplyPatchGroup")).Methods("POST", "PUT", "PATCH")
	s.Router.Handle("/auth/v1/admin/{provider}/groups/{group}", newAdminHandler(v1.DeleteGroup, "adminV1deleteGroup")).Methods("DELETE")
	s.Router.Handle("/auth/v1/admin/{provider}/groupbindings/{user}/{group}", newAdminHandler(v1.AddApplyPatchGroupBinding, "adminV1addApplyPatchGroupBinding")).Methods("POST", "PUT", "PATCH")
	s.Router.Handle("/auth/v1/admin/{provider}/groupbindings/{user}/{group}", newAdminHandler(v1.DeleteGroupBinding, "adminV1deleteGroupBinding")).Methods("DELETE")
}

func (this *Server) Start(stop <-chan struct{}) error {
	serverLog.Info("Starting Auth Server")
	certPath := filepath.Join(this.CertDir, this.CertName)
	keyPath := filepath.Join(this.CertDir, this.KeyName)

	certWatcher, err := certwatcher.New(certPath, keyPath)
	if err != nil {
		return err
	}

	go func() {
		if err := certWatcher.Start(stop); err != nil {
			serverLog.Error(err, "certificate watcher error")
		}
	}()

	cfg := &tls.Config{
		NextProtos:     []string{"h2"},
		GetCertificate: certWatcher.GetCertificate,
	}

	listener, err := tls.Listen("tcp", net.JoinHostPort(this.Host, strconv.Itoa(int(this.Port))), cfg)
	if err != nil {
		return err
	}

	serverLog.Info("serving Auth server", "host", this.Host, "port", this.Port)

	srv := &http.Server{
		Handler: this.Router,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		<-stop
		serverLog.Info("shutting down webhook server")

		// TODO: use a context with reasonable timeout
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout
			serverLog.Error(err, "error shutting down the HTTP server")
		}
		close(idleConnsClosed)
	}()

	err = srv.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	serverLog.Info("Auth Server shutdown")
	<-idleConnsClosed
	return nil
}

func serveJSON(response http.ResponseWriter, data interface{}) {
	response.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(response).Encode(data)
	if err != nil {
		panic(err)
	}
}
