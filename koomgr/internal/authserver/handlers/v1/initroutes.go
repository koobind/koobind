package v1

import (
	"github.com/gorilla/mux"
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/authserver/handlers"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/token"
	ctrl "sigs.k8s.io/controller-runtime"
)

func InitRoutes(router *mux.Router, tokenBasket token.TokenBasket, providerChain providers.ProviderChain) {

	router.Handle("/auth/v1/validateToken", &ValidateTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:      ctrl.Log.WithName(common.V1ValidateTokenUrl),
			TokenBasket: tokenBasket,
		},
	}).Methods("GET", "POST") // POST is from Api server while GET is from our client

	router.Handle("/auth/v1/getToken", &GetTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:      ctrl.Log.WithName(common.V1GetToken),
			TokenBasket: tokenBasket,
		},
		Providers: providerChain,
	}).Methods("GET")

	newAdminHandler := func(hf handlerFunc, loggerName string) *AdminV1Handler {
		return &AdminV1Handler{
			AuthHandler: handlers.AuthHandler{
				BaseHandler: handlers.BaseHandler{
					Logger:      ctrl.Log.WithName(loggerName),
					TokenBasket: tokenBasket,
				},
				Providers: providerChain,
			},
			AdminGroup:  config.Conf.AdminGroup,
			handlerFunc: hf,
		}
	}

	router.Handle("/auth/v1/admin/tokens/{token}", newAdminHandler(deleteToken, "adminV1DeleteToken")).Methods("DELETE")
	router.Handle("/auth/v1/admin/tokens", newAdminHandler(listToken, "adminV1ListToken")).Methods("GET")
	router.Handle("/auth/v1/admin/users/{user}", newAdminHandler(describeUser, "adminV1DescribeUser")).Methods("GET")

}
