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

package servers

import (
	"fmt"
	proto_v2 "github.com/koobind/koobind/koomgr/apis/proto/awh/v2"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/servers/handlers"
	handler_v2 "github.com/koobind/koobind/koomgr/internal/servers/handlers/awh/v2"
	"github.com/koobind/koobind/koomgr/internal/token"
	"github.com/koobind/koobind/koomgr/internal/token/crd"
	"github.com/koobind/koobind/koomgr/internal/token/memory"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

func Init(manager manager.Manager, kubeClient client.Client, providerChain providers.ProviderChain) {

	tokenBasket := NewTokenBasket(kubeClient)

	// Add authentication webhook Hndler in the webhook server handled by the manager
	manager.GetWebhookServer().Register(proto_v2.TokenReviewUrlPath, LogHttp(&handler_v2.TokenReviewHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:      ctrl.Log.WithName("awh-v2-tokenReview"),
			TokenBasket: tokenBasket,
		},
	}))

	if *config.Conf.AuthServer.Enabled {
		authServer := newAuthServer(tokenBasket, kubeClient, providerChain)
		err := manager.Add(authServer)
		if err != nil {
			panic(err)
		}
	}
	if *config.Conf.DexServer.Enabled {
		dexServer := newDexServer(tokenBasket, kubeClient, providerChain)
		err := manager.Add(dexServer)
		if err != nil {
			panic(err)
		}
	}

	err := manager.Add(&token.Cleaner{
		Period:      60 * time.Second,
		TokenBasket: tokenBasket,
	})
	if err != nil {
		panic(err)
	}

}

func NewTokenBasket(kubeClient client.Client) token.TokenBasket {
	if config.Conf.TokenStorage == "memory" {
		return memory.NewTokenBasket()
	} else if config.Conf.TokenStorage == "crd" {
		return crd.NewTokenBasket(kubeClient)
	} else {
		panic(fmt.Sprintf("Invalid token storage value:%s", config.Conf.TokenStorage))
	}
}
