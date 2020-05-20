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
	. "github.com/koobind/koobind/common"
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
	var user = User{Username: "Alfred"}
	userToken, err := basket.NewUserToken(user)
	assert.Nil(t, err)
	user2, ok, err := basket.Get(userToken.Token)
	assert.Nil(t, err)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")
}

func TestTimeout1(t *testing.T) {
	basket := newTokenBasket(&lifeCycle2s)
	var user = User{Username: "Alfred"}
	userToken, err := basket.NewUserToken(user)
	assert.Nil(t, err)
	time.Sleep(time.Second * 3)
	_, ok, err := basket.Get(userToken.Token)
	assert.Nil(t, err)
	assert.False(t, ok, "ok should be false")
}

func TestTimeout2(t *testing.T) {
	basket := newTokenBasket(&lifeCycle2s)
	var user = User{Username: "Alfred"}
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
