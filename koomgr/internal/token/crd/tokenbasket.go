package crd

import (
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/token"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sync"
	"time"
)

var tokenLog = ctrl.Log.WithName("token-memory")

type tokenBasket struct {
	sync.RWMutex
	defaultLifecycle *v1alpha1.TokenLifecycle
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newTokenBasket(defaultLifecycle *v1alpha1.TokenLifecycle) token.TokenBasket {
	return &tokenBasket{
		defaultLifecycle: defaultLifecycle,
	}
}

func NewTokenBasket() token.TokenBasket {
	return newTokenBasket(&v1alpha1.TokenLifecycle{
		InactivityTimeout: metav1.Duration{Duration: *config.Conf.InactivityTimeout},
		MaxTTL:            metav1.Duration{Duration: *config.Conf.SessionMaxTTL},
		ClientTTL:         metav1.Duration{Duration: *config.Conf.ClientTokenTTL},
	})
}

func (this *tokenBasket) NewUserToken(user common.User) common.UserToken {
	return common.UserToken{}
	//panic("implement me")
}

func (this *tokenBasket) Get(token string) (common.User, bool) {
	panic("implement me")
}

func (this *tokenBasket) GetAll() []common.UserToken {
	panic("implement me")
}

func (this *tokenBasket) Clean() {
	panic("implement me")
}

func (this *tokenBasket) Delete(token string) bool {
	panic("implement me")
}
