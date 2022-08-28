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
	"context"
	"fmt"
	"github.com/koobind/koobind/koomgr/apis/proto"
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/token"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"sync"
	"time"
)

var _ token.TokenBasket = &tokenBasket{}

var tokenLog = ctrl.Log.WithName("token-crd")

type tokenBasket struct {
	sync.RWMutex
	defaultLifecycle *tokenapi.TokenLifecycle
	kubeClient       client.Client
	lastHitStep      time.Duration
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newTokenBasket(kubeClient client.Client, defaultLifecycle *tokenapi.TokenLifecycle) token.TokenBasket {
	// Convert lastHitStep from % to Duration
	lhStep := (defaultLifecycle.InactivityTimeout.Duration / time.Duration(1000)) * time.Duration(config.Conf.LastHitStep)
	tokenLog.Info(fmt.Sprintf("LastHitStep:%s", lhStep.String()))
	return &tokenBasket{
		kubeClient:       kubeClient,
		defaultLifecycle: defaultLifecycle,
		lastHitStep:      lhStep,
	}
}

func NewTokenBasket(kubeClient client.Client) token.TokenBasket {
	return newTokenBasket(kubeClient, &tokenapi.TokenLifecycle{
		InactivityTimeout: metav1.Duration{Duration: *config.Conf.InactivityTimeout},
		MaxTTL:            metav1.Duration{Duration: *config.Conf.SessionMaxTTL},
		ClientTTL:         metav1.Duration{Duration: *config.Conf.ClientTokenTTL},
	})
}

func (this *tokenBasket) NewUserToken(user tokenapi.UserDesc) (proto.UserToken, error) {
	b := make([]byte, 32)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	tkn := string(b)
	now := time.Now()
	crdToken := &tokenapi.Token{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tkn,
			Namespace: config.Conf.TokenNamespace,
		},
		Spec: tokenapi.TokenSpec{
			User:      user,
			Creation:  metav1.Time{Time: now},
			Lifecycle: *this.defaultLifecycle,
		},
		Status: tokenapi.TokenStatus{
			LastHit: metav1.Time{Time: now},
		},
	}
	userToken := proto.UserToken{
		Token:   tkn,
		Spec:    tokenapi.TokenSpec{},
		LastHit: now,
	}
	crdToken.Spec.DeepCopyInto(&userToken.Spec)
	err := this.kubeClient.Create(context.TODO(), crdToken)
	if err != nil {
		tokenLog.Error(err, "token create failed", "user", user.Name)
		return proto.UserToken{}, err
	}
	tokenLog.V(0).Info("Token created", "token", tkn, "user", user.Name)

	return userToken, nil
}

func (this *tokenBasket) getToken(token string) (tokenapi.Token, bool, error) {
	crdToken := tokenapi.Token{}
	for retry := 0; retry < 4; retry++ {
		err := this.kubeClient.Get(context.TODO(), client.ObjectKey{
			Namespace: config.Conf.TokenNamespace,
			Name:      token,
		}, &crdToken)
		if err == nil {
			return crdToken, true, nil
		}
		if client.IgnoreNotFound(err) != nil {
			tokenLog.Error(err, "token Get() failed", "token", token)
			return crdToken, false, err
		}
		time.Sleep(time.Millisecond * 500)
	}
	return crdToken, false, nil // Not found is not an error. May be token has been cleaned up.
}

func (this *tokenBasket) Get(token string) (*proto.UserToken, error) {
	crdToken, found, err := this.getToken(token)
	if !found {
		return nil, err
	}
	now := time.Now()
	if stillValid(&crdToken, now) {
		err := this.touch(&crdToken, now)
		if err != nil {
			tokenLog.Error(err, "token touch on Get() failed", "token", token, "user", crdToken.Spec.User.Name)
			return nil, err
		}
		userToken := &proto.UserToken{
			Token:   token,
			Spec:    tokenapi.TokenSpec{},
			LastHit: crdToken.Status.LastHit.Time,
		}
		crdToken.Spec.DeepCopyInto(&userToken.Spec)

		return userToken, nil
	} else {
		err := this.delete(&crdToken)
		if err != nil {
			return nil, err
		}
		tokenLog.Info("Token has been cleaned on Get()", "token", token, "user", crdToken.Spec.User.Name)
		return nil, nil
	}
}

func stillValid(t *tokenapi.Token, now time.Time) bool {
	return t.Status.LastHit.Add(t.Spec.Lifecycle.InactivityTimeout.Duration).After(now) && t.Spec.Creation.Add(t.Spec.Lifecycle.MaxTTL.Duration).After(now)
}

func (this *tokenBasket) touch(t *tokenapi.Token, now time.Time) error {
	if now.After(t.Status.LastHit.Add(this.lastHitStep)) {
		tokenLog.V(1).Info("Will effectivly update LastHit", "token", t.Name, "user", t.Spec.User.Name)
		t.Status.LastHit = metav1.Time{Time: now}
		err := this.kubeClient.Update(context.TODO(), t)
		if err != nil {
			return err
		}
	} else {
		tokenLog.V(1).Info("LastHit update skipped, as too early", "token", t.Name, "user", t.Spec.User.Name)
	}
	return nil
}

func (this *tokenBasket) delete(t *tokenapi.Token) error {
	return this.kubeClient.Delete(context.TODO(), t, client.GracePeriodSeconds(0))
}

func (this *tokenBasket) GetAll() ([]proto.UserToken, error) {
	list := tokenapi.TokenList{}
	err := this.kubeClient.List(context.TODO(), &list, client.InNamespace(config.Conf.TokenNamespace))
	if err != nil {
		tokenLog.Error(err, "token List failed")
		return nil, err
	}
	slice := make([]proto.UserToken, 0, len(list.Items))
	for i := 0; i < len(list.Items); i++ {
		userToken := proto.UserToken{
			Token:   list.Items[i].Name,
			Spec:    tokenapi.TokenSpec{},
			LastHit: list.Items[i].Status.LastHit.Time,
		}
		list.Items[i].Spec.DeepCopyInto(&userToken.Spec)
		slice = append(slice, userToken)
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].Spec.Creation.Before(&slice[j].Spec.Creation)
	})
	return slice, nil
}

func (this *tokenBasket) Clean() error {
	now := time.Now()
	list := tokenapi.TokenList{}
	err := this.kubeClient.List(context.TODO(), &list, client.InNamespace(config.Conf.TokenNamespace))
	if err != nil {
		tokenLog.Error(err, "Token Cleaner. List failed")
		return err
	}
	for i := 0; i < len(list.Items); i++ {
		crdToken := list.Items[i]
		if !stillValid(&crdToken, now) {
			tokenLog.Info(fmt.Sprintf("Token %s (user:%s) has been cleaned in background.", crdToken.Name, crdToken.Spec.User.Name))
			err := this.delete(&crdToken)
			if err != nil {
				tokenLog.Error(err, "Error on delete", "token", crdToken.Name, "user", crdToken.Spec.User.Name)
				return err
			}
		}
	}
	return nil
}

func (this *tokenBasket) Delete(token string) (bool, error) {
	crdToken := tokenapi.Token{}
	err := this.kubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: config.Conf.TokenNamespace,
		Name:      token,
	}, &crdToken)
	if client.IgnoreNotFound(err) != nil {
		tokenLog.Error(err, "token Get() failed", "token", token)
		return false, err
	}
	if err != nil {
		// Token not found. Not an error (May be cleaned)
		return false, nil
	}
	err = this.delete(&crdToken)
	if err != nil {
		tokenLog.Error(err, "Error on delete", "token", crdToken.Name, "user", crdToken.Spec.User.Name)
		return false, err
	}
	return true, nil
}
