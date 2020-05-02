package memory

import (
	"fmt"
	. "github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/token"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sort"
	"sync"
	"time"
)

var tokenLog = ctrl.Log.WithName("token-memory")

func stillValid(ut *UserToken, now time.Time) bool {
	return ut.LastHit.Add(ut.Lifecycle.InactivityTimeout.Duration).After(now) && ut.Creation.Add(ut.Lifecycle.MaxTTL.Duration).After(now)
}

func touch(ut *UserToken, now time.Time) {
	ut.LastHit = now
}

type tokenBasket struct {
	sync.RWMutex
	byToken          map[string]*UserToken
	defaultLifecycle *TokenLifecycle
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newTokenBasket(defaultLifecycle *TokenLifecycle) token.TokenBasket {
	return &tokenBasket{
		byToken:          make(map[string]*UserToken),
		defaultLifecycle: defaultLifecycle,
	}
}

func NewTokenBasket() token.TokenBasket {
	return newTokenBasket(&TokenLifecycle{
		InactivityTimeout: metav1.Duration{Duration: *config.Conf.InactivityTimeout},
		MaxTTL:            metav1.Duration{Duration: *config.Conf.SessionMaxTTL},
		ClientTTL:         metav1.Duration{Duration: *config.Conf.ClientTokenTTL},
	})
}

func (this *tokenBasket) NewUserToken(user User) (UserToken, error) {
	b := make([]byte, 32)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	now := time.Now()
	t := UserToken{
		Token:     string(b),
		User:      user,
		Lifecycle: this.defaultLifecycle,
		Creation:  now,
		LastHit:   now,
	}
	this.Lock()
	this.byToken[t.Token] = &t
	this.Unlock()
	return t, nil
}

func (this *tokenBasket) Get(token string) (user User, ok bool, err error) {
	this.Lock()
	defer this.Unlock()
	ut, ok := this.byToken[token]
	if ok {
		now := time.Now()
		if stillValid(ut, now) {
			touch(ut, now)
			return ut.User, true, nil
		} else {
			delete(this.byToken, token)
			tokenLog.Info(fmt.Sprintf("Token %s (user:%s) has been cleaned on Get().", token, ut.User.Username))
			//this.log.Infof("Token %s (user:%s) has been cleaned on Get().", token, ut.User.Username)
			return User{}, false, nil
		}
	} else {
		return User{}, false, nil
	}
}

func (this *tokenBasket) GetAll() ([]UserToken, error) {
	this.RLock()
	slice := make([]UserToken, 0, len(this.byToken))
	for _, value := range this.byToken {
		slice = append(slice, *value)
	}
	this.RUnlock()
	// Stort by creation
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].Creation.Before(slice[j].Creation)
	})
	return slice, nil
}

func (this *tokenBasket) Clean() error {
	now := time.Now()
	this.Lock()
	defer this.Unlock()
	for key, value := range this.byToken {
		if !stillValid(value, now) {
			tokenLog.Info(fmt.Sprintf("Token %s (user:%s) has been cleaned in background.", key, value.User.Username))
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
