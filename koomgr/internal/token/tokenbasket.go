package token

import (
	"github.com/koobind/koobind/common"
)

type TokenBasket interface {
	NewUserToken(user common.User) (common.UserToken, error)
	Get(token string) (common.User, bool, error)
	GetAll() ([]common.UserToken, error)
	Clean() error
	Delete(token string) (bool, error)
}
