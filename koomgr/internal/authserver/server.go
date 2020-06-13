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
	"net"
	"net/http"
	"os"
	"path/filepath"
	controllerruntime "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
	"sync"
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

	Manager controllerruntime.Manager

	// defaultingOnce ensures that the default fields are only ever set once.
	defaultingOnce sync.Once

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

func (s *Server) Init() {
	s.defaultingOnce.Do(s.setDefaults)
}

// Register marks the given webhook as being served at the given path.
// It panics if two hooks are registered on the same path.
//func (s *Server) Register(path string, hook http.Handler) {
//	s.defaultingOnce.Do(s.setDefaults)
//	_, found := s.handlerByPath[path]
//	if found {
//		panic(fmt.Errorf("can't register duplicate path: %v", path))
//	}
//	// TODO(directxman12): call setfields if we've already started the server
//	s.handlerByPath[path] = hook
//	//s.Router.Handle(path, hook)
//	s.Router.PathPrefix(path).Handler(hook)
//	serverLog.Info("registering webhook", "path", path)
//}

func (this *Server) Start(stop <-chan struct{}) error {
	serverLog.Info("Starting Auth Server")
	this.defaultingOnce.Do(this.setDefaults)
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
