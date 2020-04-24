package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/token"
	"net/http"
)

type BaseHandler struct {
	Logger       logr.Logger
	TokenBasket  token.TokenBasket
	PrefixLength int
	RequestId    int
}

func (this *BaseHandler) ServeJSON(response http.ResponseWriter, data interface{}) {
	response.Header().Set("Content-Type", "application/json")
	if this.Logger.V(1).Enabled() {
		this.Logger.V(1).Info(fmt.Sprintf("Emit JSON:%s", common.JSON2String(data)))

	}
	err := json.NewEncoder(response).Encode(data)
	if err != nil {
		panic(err)
	}
}
