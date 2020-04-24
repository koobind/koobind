package token

import (
	. "github.com/koobind/koobind/common"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

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
	userToken := basket.NewUserToken(user)
	user2, ok := basket.Get(userToken.Token)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")
}

func TestTimeout1(t *testing.T) {
	basket := newTokenBasket(&lifeCycle2s)
	var user = User{Username: "Alfred"}
	token := basket.NewUserToken(user).Token
	time.Sleep(time.Second * 3)
	_, ok := basket.Get(token)
	assert.False(t, ok, "ok should be false")
}

func TestTimeout2(t *testing.T) {
	basket := newTokenBasket(&lifeCycle2s)
	var user = User{Username: "Alfred"}
	token := basket.NewUserToken(user).Token

	time.Sleep(time.Second)

	user2, ok := basket.Get(token)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")

	time.Sleep(time.Second)

	user2, ok = basket.Get(token)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")

	time.Sleep(time.Second)

	user2, ok = basket.Get(token)
	assert.True(t, ok, "ok should be true")
	assert.Equal(t, "Alfred", user2.Username, "User should be Alfred")

	time.Sleep(time.Second * 3)

	user2, ok = basket.Get(token)
	assert.False(t, ok, "ok should be false")
}
