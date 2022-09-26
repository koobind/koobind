package servers

import (
	"github.com/koobind/koobind/koomgr/apis/proto"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/servers/handlers"
	v1 "github.com/koobind/koobind/koomgr/internal/servers/handlers/v1"
	"github.com/koobind/koobind/koomgr/internal/token"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func newDexServer(tokenBasket token.TokenBasket, kubeClient client.Client, providerChain providers.ProviderChain) *Server {
	s := &Server{
		Name:    "dex",
		Logger:  logf.Log.WithName("httpserver").WithValues("name", "dex"),
		Host:    config.Conf.DexServer.Host,
		Port:    config.Conf.DexServer.Port,
		CertDir: config.Conf.DexServer.CertDir,
		NoSsl:   *config.Conf.DexServer.NoSsl,
	}
	s.setDefaults()
	s.Router.Handle(proto.DexLoginUrlPath, &v1.DexLoginHandler{
		BaseHandler: handlers.BaseHandler{
			Logger:      ctrlrt.Log.WithName("authV1validateToken"),
			TokenBasket: tokenBasket,
		},
		Providers: providerChain,
	}).Methods("POST")
	return s
}
