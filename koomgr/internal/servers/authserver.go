package servers

import (
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/servers/handlers"
	v1 "github.com/koobind/koobind/koomgr/internal/servers/handlers/v1"
	"github.com/koobind/koobind/koomgr/internal/token"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func newAuthServer(tokenBasket token.TokenBasket, kubeClient client.Client, providerChain providers.ProviderChain) *Server {
	s := &Server{
		Name:    "auth",
		Logger:  logf.Log.WithName("httpserver").WithValues("name", "auth"),
		Host:    config.Conf.AuthServer.Host,
		Port:    config.Conf.AuthServer.Port,
		CertDir: config.Conf.AuthServer.CertDir,
		NoSsl:   *config.Conf.AuthServer.NoSsl,
	}
	s.setDefaults()
	s.Router.Handle("/auth/v1/validateToken", &v1.ValidateTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:      ctrlrt.Log.WithName("authV1validateToken"),
			TokenBasket: tokenBasket,
		},
	}).Methods("GET", "POST") // POST is from Api server while GET is from our client

	s.Router.Handle("/auth/v1/getToken", &v1.GetTokenHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:      ctrlrt.Log.WithName("v1getToken"),
			TokenBasket: tokenBasket,
		},
		Providers: providerChain,
	}).Methods("GET")

	s.Router.Handle("/auth/v1/changePassword", &v1.ChangePasswordHandler{
		AuthHandler: handlers.AuthHandler{
			BaseHandler: handlers.BaseHandler{
				Logger:      ctrlrt.Log.WithName("v1changePassword"),
				TokenBasket: tokenBasket,
			},
			Providers: providerChain,
		},
	}).Methods("POST")

	newAdminHandler := func(hf v1.HandlerFunc, loggerName string) *v1.AdminV1Handler {
		return &v1.AdminV1Handler{
			AuthHandler: handlers.AuthHandler{
				BaseHandler: handlers.BaseHandler{
					Logger:      ctrlrt.Log.WithName(loggerName),
					TokenBasket: tokenBasket,
				},
				Providers: providerChain,
			},
			AdminGroup:  config.Conf.AdminGroup,
			KubeClient:  kubeClient,
			HandlerFunc: hf,
		}
	}

	s.Router.Handle("/auth/v1/admin/tokens/{token}", newAdminHandler(v1.DeleteToken, "adminV1deleteToken")).Methods("DELETE")
	s.Router.Handle("/auth/v1/admin/tokens", newAdminHandler(v1.ListToken, "adminV1listToken")).Methods("GET")
	s.Router.Handle("/auth/v1/admin/users/{user}", newAdminHandler(v1.DescribeUser, "adminV1describeUser")).Methods("GET")
	s.Router.Handle("/auth/v1/admin/{provider}/users/{user}", newAdminHandler(v1.AddApplyPatchUser, "adminV1addApplyPatchUser")).Methods("POST", "PUT", "PATCH")
	s.Router.Handle("/auth/v1/admin/{provider}/users/{user}", newAdminHandler(v1.DeleteUser, "adminV1deleteUser")).Methods("DELETE")
	s.Router.Handle("/auth/v1/admin/{provider}/groups/{group}", newAdminHandler(v1.AddApplyPatchGroup, "adminV1addApplyPatchGroup")).Methods("POST", "PUT", "PATCH")
	s.Router.Handle("/auth/v1/admin/{provider}/groups/{group}", newAdminHandler(v1.DeleteGroup, "adminV1deleteGroup")).Methods("DELETE")
	s.Router.Handle("/auth/v1/admin/{provider}/groupbindings/{user}/{group}", newAdminHandler(v1.AddApplyPatchGroupBinding, "adminV1addApplyPatchGroupBinding")).Methods("POST", "PUT", "PATCH")
	s.Router.Handle("/auth/v1/admin/{provider}/groupbindings/{user}/{group}", newAdminHandler(v1.DeleteGroupBinding, "adminV1deleteGroupBinding")).Methods("DELETE")
	return s
}
