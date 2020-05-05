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
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	crtzap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	directoryv1alpha1 "github.com/koobind/koobind/koomgr/apis/directory/v1alpha1"
	tokensv1alpha1 "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/authserver"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers/chain"
	"github.com/koobind/koobind/koomgr/internal/token"
	"github.com/koobind/koobind/koomgr/internal/token/crd"
	"github.com/koobind/koobind/koomgr/internal/token/memory"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = directoryv1alpha1.AddToScheme(scheme)
	_ = tokensv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {

	config.Setup()

	ll := zap.NewAtomicLevelAt(zapcore.Level(-config.Conf.LogLevel))
	stLevel := zap.NewAtomicLevelAt(zapcore.Level(zapcore.DPanicLevel)) // No stack trace for WARN and ERROR
	ctrl.SetLogger(crtzap.New(crtzap.UseDevMode(config.Conf.LogMode == "dev"), crtzap.Level(&ll), crtzap.StacktraceLevel(&stLevel)))

	setupLog.V(1).Info("Debug log mode activated")
	setupLog.V(2).Info(
		"Trace log mode activated")
	setupLog.V(3).Info("Verbose trace log mode activated")
	setupLog.V(4).Info("Very verbose trace log mode activated")

	// Must be BEFORE manager creation, to have accurate list of Namespace
	providerChain, err := chain.BuildProviderChain(&config.Conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	namespaceList := config.Conf.CrdNamespaces.DeepCopy().Add(config.Conf.TokenNamespace).AsList()
	//setupLog.Info(fmt.Sprintf("Namespaces handled by Kube client cache:%v (For webhook:%v)", namespaceList, config.Conf.CrdNamespaces.AsList()))
	setupLog.Info("Namespaces", "kubeClient", namespaceList, "webhook", config.Conf.CrdNamespaces.AsList())

	cfg := ctrl.GetConfigOrDie()
	setupLog.V(2).Info(fmt.Sprintf("config:%v", cfg))
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		LeaderElection:     false,
		Port:               config.Conf.WebhookServer.Port,
		CertDir:            config.Conf.WebhookServer.CertDir,
		Host:               config.Conf.WebhookServer.Host,
		NewCache:           cache.MultiNamespacedCacheBuilder(namespaceList),
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&directoryv1alpha1.User{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "User")
		os.Exit(1)
	}
	if err = (&directoryv1alpha1.Group{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Group")
		os.Exit(1)
	}
	if err = (&directoryv1alpha1.GroupBinding{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "GroupBinding")
		os.Exit(1)
	}

	authserver.Init(mgr, NewTokenBasket(mgr.GetClient()), providerChain)

	err = mgr.GetFieldIndexer().IndexField(context.TODO(), &directoryv1alpha1.GroupBinding{}, "userkey", func(rawObj runtime.Object) []string {
		ugb := rawObj.(*directoryv1alpha1.GroupBinding)
		return []string{ugb.Spec.User}
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Set global KubeClient
	config.KubeClient = mgr.GetClient()

	// +kubebuilder:scaffold:builder
	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func NewTokenBasket(kubeClient client.Client) token.TokenBasket {
	if config.Conf.TokenStorage == "memory" {
		return memory.NewTokenBasket()
	} else if config.Conf.TokenStorage == "crd" {
		return crd.NewTokenBasket(kubeClient)
	} else {
		panic(fmt.Sprintf("Invalid token storage value:%s", config.Conf.TokenStorage))
	}
}
