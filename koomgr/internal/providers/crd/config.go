package crd

import (
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CrdProviderConfig struct {
	config.BaseProviderConfig `yaml:",inline"`
}

func (this *CrdProviderConfig) Open(idx int, configFolder string, kubeClient client.Client) (providers.Provider, error) {
	if err := this.InitBase(idx); err != nil {
		return nil, err
	}
	prvd := crdProvider{
		CrdProviderConfig: this,
		kubeClient:        kubeClient,
	}
	return &prvd, nil
}
