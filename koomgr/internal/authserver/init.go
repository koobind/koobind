package authserver

import (
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/authserver/handlers"
	"github.com/koobind/koobind/koomgr/internal/authserver/handlers/v1"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/token"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func Init(manager manager.Manager, tokenBasket token.TokenBasket, providerChain providers.ProviderChain) {

	manager.GetWebhookServer().Register(common.V1ValidateTokenUrl, LogHttp(&v1.ValidateTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:       ctrl.Log.WithName(common.V1ValidateTokenUrl),
			TokenBasket:  tokenBasket,
			PrefixLength: len(common.V1ValidateTokenUrl),
		},
	}))
	manager.GetWebhookServer().Register(common.V1GetToken, LogHttp(&v1.GetTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:       ctrl.Log.WithName(common.V1GetToken),
			TokenBasket:  tokenBasket,
			PrefixLength: len(common.V1GetToken),
		},
		Providers: providerChain,
	}))
	manager.GetWebhookServer().Register(common.V1Admin, LogHttp(&v1.AdminV1Handler{
		AuthHandler: handlers.AuthHandler{
			BaseHandler: handlers.BaseHandler{
				Logger:       ctrl.Log.WithName(common.V1Admin),
				TokenBasket:  tokenBasket,
				PrefixLength: len(common.V1Admin),
			},
			Providers: providerChain,
		},
		AdminGroup: config.Conf.AdminGroup,
	}))

}
