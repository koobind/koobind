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
	"time"
)

func Init(manager manager.Manager, tokenBasket token.TokenBasket, providerChain providers.ProviderChain) {

	// There is two endpoint:
	// - Webhook server, handling all handlerByPath (Mutating, validating an authitication). Called only by API server
	// - Auth server, handling all requests from koocli. Exposed externally by a nodeport
	// ValidateTokenHandler is set on both, as will be called from API server (POST) and koocli (GET)

	manager.GetWebhookServer().Register(common.V1ValidateTokenUrl, LogHttp(&v1.ValidateTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:       ctrl.Log.WithName(common.V1ValidateTokenUrl),
			TokenBasket:  tokenBasket,
			PrefixLength: len(common.V1ValidateTokenUrl),
		},
	}))

	authServer := Server{
		Host:    config.Conf.AuthServer.Host,
		Port:    config.Conf.AuthServer.Port,
		CertDir: config.Conf.AuthServer.CertDir,
	}
	err := manager.Add(&authServer)
	if err != nil {
		panic(err)
	}

	authServer.Register("/auth/v1/validateToken", &v1.ValidateTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:       ctrl.Log.WithName(common.V1ValidateTokenUrl),
			TokenBasket:  tokenBasket,
			PrefixLength: len(common.V1ValidateTokenUrl),
		},
	})
	authServer.Register("/auth/v1/getToken", &v1.GetTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:       ctrl.Log.WithName(common.V1GetToken),
			TokenBasket:  tokenBasket,
			PrefixLength: len(common.V1GetToken),
		},
		Providers: providerChain,
	})
	authServer.Register("/auth/v1/admin/", &v1.AdminV1Handler{
		AuthHandler: handlers.AuthHandler{
			BaseHandler: handlers.BaseHandler{
				Logger:       ctrl.Log.WithName(common.V1Admin),
				TokenBasket:  tokenBasket,
				PrefixLength: len(common.V1Admin),
			},
			Providers: providerChain,
		},
		AdminGroup: config.Conf.AdminGroup,
	})

	err = manager.Add(&Cleaner{
		Period:      60 * time.Second,
		TokenBasket: tokenBasket,
	})
	if err != nil {
		panic(err)
	}

}
