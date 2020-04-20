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

package main

import (
	"flag"
	"fmt"
	directoryv1alpha1 "github.com/koobind/koobind/koomgr/apis/directory/v1alpha1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	crtzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = directoryv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var managerCertDir string
	var logLevel int
	var host string
	flag.StringVar(&host, "host", "", "Webhook server bind address")
	flag.StringVar(&managerCertDir, "cert-dir", "", "Path to the server certificate folder")
	flag.StringVar(&directoryv1alpha1.Namespace, "namespace", "", "The namespace where to store koo resources (users,groups,bindings)")
	flag.IntVar(&logLevel, "logLevel", 0, "Log level (0:INFO; 1:DEBUG, 2:MoreDebug...)")
	flag.Parse()

	if logLevel > 0 {
		ll := zap.NewAtomicLevelAt(zapcore.Level(-logLevel))
		ctrl.SetLogger(crtzap.New(crtzap.UseDevMode(true), crtzap.Level(&ll)))
	} else {
		ctrl.SetLogger(crtzap.New())
	}

	setupLog.V(1).Info("Debug log mode activated")
	setupLog.V(2).Info("Trace log mode activated")
	setupLog.V(3).Info("Verbose trace log mode activated")
	setupLog.V(4).Info("Very verbose trace log mode activated")

	if directoryv1alpha1.Namespace == "" {
		fmt.Fprintf(os.Stderr, "ERROR: --namespace parameter is required!\n")
		os.Exit(2)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		Port:               9443,
		LeaderElection:     false,
		//LeaderElectionID:   "f9553f09.koobind.io",
		CertDir: managerCertDir,
		Host:    host,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&directoryv1alpha1.User{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "User")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
