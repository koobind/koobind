package authserver

import (
	"github.com/koobind/koobind/koomgr/internal/authserver/handlers"
	"github.com/koobind/koobind/koomgr/internal/authserver/handlers/v1"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/token"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	v1ValidateTokenUrl = "/auth/v1/validateToken"
	v1GetToken         = "/auth/v1/getToken"
	v1Admin            = "/auth/v1/admin/"
)

func Init(manager manager.Manager, tokenBasket token.TokenBasket, providerChain providers.ProviderChain) {

	manager.GetWebhookServer().Register(v1ValidateTokenUrl, LogHttp(&v1.ValidateTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:       ctrl.Log.WithName(v1ValidateTokenUrl),
			TokenBasket:  tokenBasket,
			PrefixLength: len(v1ValidateTokenUrl),
		},
	}))
	manager.GetWebhookServer().Register(v1GetToken, LogHttp(&v1.GetTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:       ctrl.Log.WithName(v1GetToken),
			TokenBasket:  tokenBasket,
			PrefixLength: len(v1GetToken),
		},
		Providers: providerChain,
	}))
	manager.GetWebhookServer().Register(v1Admin, LogHttp(&v1.AdminV1Handler{
		AuthHandler: handlers.AuthHandler{
			BaseHandler: handlers.BaseHandler{
				Logger:       ctrl.Log.WithName(v1Admin),
				TokenBasket:  tokenBasket,
				PrefixLength: len(v1Admin),
			},
			Providers: providerChain,
		},
		AdminGroup: config.Conf.AdminGroup,
	}))

}
