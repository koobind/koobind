package crd

import (
	"fmt"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CrdProviderConfig struct {
	config.BaseProviderConfig `yaml:",inline"`
	Namespace                 string `yaml:"namespace"` // The namespace where koo resources (users,groups,groupbindings) are stored
}

func (this *CrdProviderConfig) Open(idx int, configFolder string) (providers.Provider, error) {
	if err := this.InitBase(idx); err != nil {
		return nil, err
	}
	prvd := crdProvider{
		CrdProviderConfig: this,
	}
	if prvd.Namespace == "" {
		if config.Conf.Namespace == "" {
			return &prvd, fmt.Errorf("Missing providers.%s.namespace and no global 'namespace' parameter provided", prvd.Name)
		}
		prvd.Namespace = config.Conf.Namespace
	}
	prvd.logger = ctrl.Log.WithName("crd:" + prvd.Name)
	// Add this namespace as valid ones for the validating webhooks.
	config.Conf.CrdNamespaces.Add(prvd.Namespace)
	return &prvd, nil
}
