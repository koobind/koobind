package token

import (
	"github.com/koobind/koobind/common"
)

type TokenBasket interface {
	NewUserToken(user common.User) common.UserToken
	Get(token string) (common.User, bool)
	GetAll() []common.UserToken
	Clean()
	Delete(token string) bool
}
