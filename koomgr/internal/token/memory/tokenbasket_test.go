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
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

var lifeCycle2s = tokenapi.TokenLifecycle{
	InactivityTimeout: ParseDurationOrPanic("2s"),
	MaxTTL:            ParseDurationOrPanic("24h"),
	ClientTTL:         ParseDurationOrPanic("10s"),
}

var lifeCycle3s = tokenapi.TokenLifecycle{
	InactivityTimeout: ParseDurationOrPanic("3s"),
	MaxTTL:            ParseDurationOrPanic("24h"),
	ClientTTL:         ParseDurationOrPanic("10s"),
}

//func init() {
//
//	lifeCycle3s = TokenLifecycle{
//		InactivityTimeout: ParseDurationOrPanic("3s"),
//		MaxTTL:            ParseDurationOrPanic("24h"),
//		ClientTTL:         ParseDurationOrPanic("10s"),
//	}
//
//}

func TestNew(t *testing.T) {
	basket := newTokenBasket(&lifeCycle3s)
	var user = tokenapi.UserDesc{Name: "Alfred", Groups: []string{}}
	userToken, err := basket.NewUserToken(user)
	assert.Nil(t, err)
	userToken2, err := basket.Get(userToken.Token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken should be found")
	assert.Equal(t, "Alfred", userToken2.Spec.User.Name, "User should be Alfred")
}

func TestTimeout1(t *testing.T) {
	basket := newTokenBasket(&lifeCycle2s)
	var user = tokenapi.UserDesc{Name: "Alfred", Groups: []string{}}
	userToken, err := basket.NewUserToken(user)
	assert.Nil(t, err)
	time.Sleep(time.Second * 3)
	userToken2, err := basket.Get(userToken.Token)
	assert.Nil(t, err)
	assert.Nil(t, userToken2, "userToken should be nil (Not found)")
}

func TestTimeout2(t *testing.T) {
	basket := newTokenBasket(&lifeCycle2s)
	var user = tokenapi.UserDesc{Name: "Alfred", Groups: []string{}}
	userToken, err := basket.NewUserToken(user)
	assert.Nil(t, err)
	token := userToken.Token

	time.Sleep(time.Second)

	userToken2, err := basket.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken2 should be found")
	assert.Equal(t, "Alfred", userToken2.Spec.User.Name, "User should be Alfred")

	time.Sleep(time.Second)

	userToken2, err = basket.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken2 should be found")
	assert.Equal(t, "Alfred", userToken2.Spec.User.Name, "User should be Alfred")

	time.Sleep(time.Second)

	userToken2, err = basket.Get(token)
	assert.Nil(t, err)
	assert.NotNil(t, userToken2, "userToken2 should be found")
	assert.Equal(t, "Alfred", userToken2.Spec.User.Name, "User should be Alfred")

	time.Sleep(time.Second * 3)

	userToken2, err = basket.Get(token)
	assert.Nil(t, err)
	assert.Nil(t, userToken2, "userToken2 shound not be found")
}
