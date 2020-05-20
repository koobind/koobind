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
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
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

var tokenLog = ctrl.Log.WithName("token-crd")

type tokenBasket struct {
	sync.RWMutex
	defaultLifecycle *common.TokenLifecycle
	kubeClient       client.Client
	lastHitStep      time.Duration
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newTokenBasket(kubeClient client.Client, defaultLifecycle *common.TokenLifecycle) token.TokenBasket {
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
	return newTokenBasket(kubeClient, &common.TokenLifecycle{
		InactivityTimeout: metav1.Duration{Duration: *config.Conf.InactivityTimeout},
		MaxTTL:            metav1.Duration{Duration: *config.Conf.SessionMaxTTL},
		ClientTTL:         metav1.Duration{Duration: *config.Conf.ClientTokenTTL},
	})
}

func (this *tokenBasket) NewUserToken(user common.User) (common.UserToken, error) {
	b := make([]byte, 32)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	now := time.Now()
	t := common.UserToken{
		Token:     string(b),
		User:      user,
		Lifecycle: this.defaultLifecycle,
		Creation:  now,
		LastHit:   now,
	}
	crdToken := &v1alpha1.Token{
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.Token,
			Namespace: config.Conf.TokenNamespace,
		},
		Spec: v1alpha1.TokenSpec{
			User:      user,
			Creation:  metav1.Time{Time: now},
			Lifecycle: *this.defaultLifecycle,
		},
		Status: v1alpha1.TokenStatus{
			LastHit: metav1.Time{Time: now},
		},
	}
	err := this.kubeClient.Create(context.TODO(), crdToken)
	if err != nil {
		tokenLog.Error(err, "token create failed", "user", user.Username)
		return common.UserToken{}, err
	}
	tokenLog.V(0).Info("Token created", "token", t.Token, "user", user.Username)
	return t, nil
}

func (this *tokenBasket) Get(token string) (common.User, bool, error) {
	crdToken := v1alpha1.Token{}
	err := this.kubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: config.Conf.TokenNamespace,
		Name:      token,
	}, &crdToken)
	if client.IgnoreNotFound(err) != nil {
		tokenLog.Error(err, "token Get() failed", "token", token)
		return common.User{}, false, err
	}
	if err != nil {
		// Token not found. Not an error (May be cleaned)
		return common.User{}, false, nil
	}
	now := time.Now()
	if stillValid(&crdToken, now) {
		err := this.touch(&crdToken, now)
		if err != nil {
			tokenLog.Error(err, "token touch on Get() failed", "token", token, "user", crdToken.Spec.User.Username)
			return common.User{}, false, err
		}
		return crdToken.Spec.User, true, nil
	} else {
		err := this.delete(&crdToken)
		if err != nil {
			return common.User{}, false, nil
		}
		tokenLog.Info("Token has been cleaned on Get()", "token", token, "user", crdToken.Spec.User.Username)
		return common.User{}, false, nil
	}
}

func stillValid(t *v1alpha1.Token, now time.Time) bool {
	return t.Status.LastHit.Add(t.Spec.Lifecycle.InactivityTimeout.Duration).After(now) && t.Spec.Creation.Add(t.Spec.Lifecycle.MaxTTL.Duration).After(now)
}

func (this *tokenBasket) touch(t *v1alpha1.Token, now time.Time) error {
	if now.After(t.Status.LastHit.Add(this.lastHitStep)) {
		tokenLog.V(1).Info("Will effectivly update LastHit", "token", t.Name, "user", t.Spec.User.Username)
		t.Status.LastHit = metav1.Time{Time: now}
		err := this.kubeClient.Update(context.TODO(), t)
		if err != nil {
			return err
		}
	} else {
		tokenLog.V(1).Info("LastHit update skipped, as too early", "token", t.Name, "user", t.Spec.User.Username)
	}
	return nil
}

func (this *tokenBasket) delete(t *v1alpha1.Token) error {
	return this.kubeClient.Delete(context.TODO(), t, client.GracePeriodSeconds(0))
}

func (this *tokenBasket) GetAll() ([]common.UserToken, error) {
	list := v1alpha1.TokenList{}
	err := this.kubeClient.List(context.TODO(), &list, client.InNamespace(config.Conf.TokenNamespace))
	if err != nil {
		tokenLog.Error(err, "token List failed")
		return []common.UserToken{}, err
	}
	slice := make([]common.UserToken, 0, len(list.Items))
	for i := 0; i < len(list.Items); i++ {
		crdToken := list.Items[i]
		slice = append(slice, common.UserToken{
			Token:     crdToken.Name,
			User:      crdToken.Spec.User,
			Creation:  crdToken.Spec.Creation.Time,
			LastHit:   crdToken.Status.LastHit.Time,
			Lifecycle: &crdToken.Spec.Lifecycle,
		})

	}
	// Stort by creation
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].Creation.Before(slice[j].Creation)
	})
	return slice, nil
}

func (this *tokenBasket) Clean() error {
	now := time.Now()
	list := v1alpha1.TokenList{}
	err := this.kubeClient.List(context.TODO(), &list, client.InNamespace(config.Conf.TokenNamespace))
	if err != nil {
		tokenLog.Error(err, "Token Cleaner. List failed")
		return err
	}
	for i := 0; i < len(list.Items); i++ {
		crdToken := list.Items[i]
		if !stillValid(&crdToken, now) {
			tokenLog.Info(fmt.Sprintf("Token %s (user:%s) has been cleaned in background.", crdToken.Name, crdToken.Spec.User.Username))
			err := this.delete(&crdToken)
			if err != nil {
				tokenLog.Error(err, "Error on delete", "token", crdToken.Name, "user", crdToken.Spec.User.Username)
				return err
			}
		}
	}
	return nil
}

func (this *tokenBasket) Delete(token string) (bool, error) {
	crdToken := v1alpha1.Token{}
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
		tokenLog.Error(err, "Error on delete", "token", crdToken.Name, "user", crdToken.Spec.User.Username)
		return false, err
	}
	return true, nil
}
