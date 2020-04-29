package static

import (
	"fmt"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type User struct {
	Login        string   `yaml:"login"`
	PasswordHash string   `yaml:"passwordHash"`
	Id           *int     `yaml:"id,omitempty"`
	Groups       []string `yaml:"groups"`
	Email        string   `yaml:"email"`
}

type StaticProviderConfig struct {
	config.BaseProviderConfig `yaml:",inline"`
	Users                     []User `yaml:"users"`
}

func (this *StaticProviderConfig) Open(idx int, configFolder string, kubeClient client.Client) (providers.Provider, error) {
	if err := this.InitBase(idx); err != nil {
		return nil, err
	}
	prvd := staticProvider{
		StaticProviderConfig: this,
		userByLogin:          make(map[string]User),
	}
	for _, u := range this.Users {
		if _, exists := prvd.userByLogin[u.Login]; exists {
			return nil, fmt.Errorf("user '%s' is defined twice in the static provider '%s'", u.Login, this.Name)
		} else {
			prvd.userByLogin[u.Login] = u
		}
	}
	return &prvd, nil
}
