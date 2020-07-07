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
package crd

import (
	. "github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crtzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
	"time"
)

func ParseDuration(d string) (metav1.Duration, error) {
	duration, err := time.ParseDuration(d)
	if err == nil {
		return metav1.Duration{Duration: duration}, nil
	} else {
		return metav1.Duration{}, err
	}
}

func ParseDurationOrPanic(d string) metav1.Duration {
	duration, err := ParseDuration(d)
	if err != nil {
		panic(err)
	}
	return duration
}

var lifeCycle2s TokenLifecycle = TokenLifecycle{
	InactivityTimeout: ParseDurationOrPanic("2s"),
	MaxTTL:            ParseDurationOrPanic("24h"),
	ClientTTL:         ParseDurationOrPanic("10s"),
}

var lifeCycle3s TokenLifecycle = TokenLifecycle{
	InactivityTimeout: ParseDurationOrPanic("3s"),
	MaxTTL:            ParseDurationOrPanic("24h"),
	ClientTTL:         ParseDurationOrPanic("10s"),
}

func newClient() client.Client {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = "~/.kube/config"
	}
	myconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	crScheme := runtime.NewScheme()
	err = v1alpha1.AddToScheme(crScheme)
	if err != nil {
		panic(err)
	}
	myclient, err := client.New(myconfig, client.Options{
		Scheme: crScheme,
	})
	if err != nil {
		panic(err)
	}
	return myclient
}

func TestMain(m *testing.M) {
	config.Conf.TokenNamespace = "koo-system"
	config.Conf.LastHitStep = 100
	config.Conf.LogLevel = -1

	ll := zap.NewAtomicLevelAt(zapcore.Level(-config.Conf.LogLevel))
	stLevel := zap.NewAtomicLevelAt(zapcore.Level(zapcore.DPanicLevel)) // No stack trace for WARN and ERROR
	ctrl.SetLogger(crtzap.New(crtzap.UseDevMode(config.Conf.LogMode == "dev"), crtzap.Level(&ll), crtzap.StacktraceLevel(&stLevel)))

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	basket := newTokenBasket(newClient(), &lifeCycle3s)
	var user = User{Username: "Alfred", Groups: []string{}}
	userToken, err := basket.NewUserToken(user)
	assert.Nil(t, err)
	user2, ok, err := basket.Get(userToken.Token)
	assert.Nil(t, err)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")
}

func TestTimeout1(t *testing.T) {
	basket := newTokenBasket(newClient(), &lifeCycle2s)
	var user = User{Username: "Alfred", Groups: []string{}}
	userToken, err := basket.NewUserToken(user)
	assert.Nil(t, err)
	time.Sleep(time.Second * 3)
	_, ok, err := basket.Get(userToken.Token)
	assert.Nil(t, err)
	assert.False(t, ok, "ok should be false")
}

func TestTimeout2(t *testing.T) {
	basket := newTokenBasket(newClient(), &lifeCycle2s)
	var user = User{Username: "Alfred", Groups: []string{}}
	userToken, err := basket.NewUserToken(user)
	assert.Nil(t, err)
	token := userToken.Token

	time.Sleep(time.Second)

	user2, ok, err := basket.Get(token)
	assert.Nil(t, err)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")

	time.Sleep(time.Second)

	user2, ok, err = basket.Get(token)
	assert.Nil(t, err)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")

	time.Sleep(time.Second)

	user2, ok, err = basket.Get(token)
	assert.Nil(t, err)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")

	time.Sleep(time.Second * 3)

	user2, ok, err = basket.Get(token)
	assert.Nil(t, err)
	assert.False(t, ok, "ok should be false")
}

func TestMultipleGet(t *testing.T) {
	basket := newTokenBasket(newClient(), &lifeCycle3s)
	var user = User{Username: "Alfred", Groups: []string{}}
	userToken, err := basket.NewUserToken(user)
	assert.Nil(t, err)

	user2, ok, err := basket.Get(userToken.Token)
	assert.Nil(t, err)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")

	time.Sleep(time.Second)
	user2, ok, err = basket.Get(userToken.Token)
	assert.Nil(t, err)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")

	user2, ok, err = basket.Get(userToken.Token)
	assert.Nil(t, err)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")
}

func TestMultipleBasket(t *testing.T) {
	basket1 := newTokenBasket(newClient(), &lifeCycle3s)
	basket2 := newTokenBasket(newClient(), &lifeCycle3s)
	var user = User{Username: "Alfred", Groups: []string{}}
	userToken, err := basket1.NewUserToken(user)
	assert.Nil(t, err)

	user2, ok, err := basket1.Get(userToken.Token)
	assert.Nil(t, err)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")

	//time.Sleep(time.Second * 2)
	user2, ok, err = basket2.Get(userToken.Token)
	assert.Nil(t, err)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")

	time.Sleep(time.Second)
	user2, ok, err = basket1.Get(userToken.Token)
	assert.Nil(t, err)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")
}
