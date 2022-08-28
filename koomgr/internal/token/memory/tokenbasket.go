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

package memory

import (
	"fmt"
	"github.com/koobind/koobind/koomgr/apis/proto"
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/token"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sort"
	"sync"
	"time"
)

var _ token.TokenBasket = &tokenBasket{}

var tokenLog = ctrl.Log.WithName("token-memory")

func stillValid(ut *proto.UserToken, now time.Time) bool {
	return ut.LastHit.Add(ut.Spec.Lifecycle.InactivityTimeout.Duration).After(now) && ut.Spec.Creation.Add(ut.Spec.Lifecycle.MaxTTL.Duration).After(now)
}

func touch(ut *proto.UserToken, now time.Time) {
	ut.LastHit = now
}

type tokenBasket struct {
	sync.RWMutex
	byToken          map[string]*proto.UserToken
	defaultLifecycle *tokenapi.TokenLifecycle
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newTokenBasket(defaultLifecycle *tokenapi.TokenLifecycle) token.TokenBasket {
	return &tokenBasket{
		byToken:          make(map[string]*proto.UserToken),
		defaultLifecycle: defaultLifecycle,
	}
}

func NewTokenBasket() token.TokenBasket {
	return newTokenBasket(&tokenapi.TokenLifecycle{
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
	now := time.Now()
	t := proto.UserToken{
		Token: string(b),
		Spec: tokenapi.TokenSpec{
			User:      user,
			Creation:  metav1.Time{Time: now},
			Lifecycle: *this.defaultLifecycle,
		},
		LastHit: now,
	}
	this.Lock()
	this.byToken[t.Token] = &t
	this.Unlock()
	return t, nil
}

func (this *tokenBasket) Get(token string) (userToken *proto.UserToken, err error) {
	this.Lock()
	defer this.Unlock()
	ut, ok := this.byToken[token]
	if ok {
		now := time.Now()
		if stillValid(ut, now) {
			touch(ut, now)
			return ut, nil
		} else {
			delete(this.byToken, token)
			tokenLog.Info(fmt.Sprintf("Token %s (user:%s) has been cleaned on Get().", token, ut.Spec.User.Name))
			//this.log.Infof("Token %s (user:%s) has been cleaned on Get().", token, ut.User.Username)
			return nil, nil
		}
	} else {
		return nil, nil
	}
}

func (this *tokenBasket) GetAll() ([]proto.UserToken, error) {
	this.RLock()
	slice := make([]proto.UserToken, 0, len(this.byToken))
	for _, value := range this.byToken {
		slice = append(slice, *value)
	}
	this.RUnlock()
	// Stort by creation
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].Spec.Creation.Before(&slice[j].Spec.Creation)
	})
	return slice, nil
}

func (this *tokenBasket) Clean() error {
	now := time.Now()
	this.Lock()
	defer this.Unlock()
	for key, value := range this.byToken {
		if !stillValid(value, now) {
			tokenLog.Info(fmt.Sprintf("Token %s (user:%s) has been cleaned in background.", key, value.Spec.User.Name))
			//this.log.Infof("Token %s (user:%s) has been cleaned in background.", key, value.User.Username)
			delete(this.byToken, key)
		}
	}
	return nil
}

// Return true if there was a token to delete
func (this *tokenBasket) Delete(token string) (bool, error) {
	this.Lock()
	defer this.Unlock()
	_, ok := this.byToken[token]
	if ok {
		delete(this.byToken, token)
	}
	return ok, nil
}
