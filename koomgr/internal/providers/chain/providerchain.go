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
package chain

import (
	"fmt"
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/providers/crd"
	"github.com/koobind/koobind/koomgr/internal/providers/ldap"
	"github.com/koobind/koobind/koomgr/internal/providers/static"
	"gopkg.in/yaml.v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sort"
)

type providerChain struct {
	providers      []providers.Provider
	prividerByName map[string]providers.Provider
}

var pcLog = ctrl.Log.WithName("providerChain")

type providerConfig interface {
	Open(idx int, configFolder string) (providers.Provider, error)
	GetName() string
	IsEnabled() bool
}

var ProviderConfigBuilderFromType = map[string]func() providerConfig{
	"static": func() providerConfig { return new(static.StaticProviderConfig) },
	"ldap":   func() providerConfig { return new(ldap.LdapProviderConfig) },
	"crd":    func() providerConfig { return new(crd.CrdProviderConfig) },
}

func BuildProviderChain(conf *config.Config) (providers.ProviderChain, error) {
	this := providerChain{
		providers:      []providers.Provider{},
		prividerByName: make(map[string]providers.Provider),
	}
	for i := 0; i < len(conf.Providers); i++ {
		//var m map[interface{}]interface{}
		m, ok := conf.Providers[i].(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("provider[%d] is not a map", i)
		}
		t, ok := m["type"]
		if !ok {
			return nil, fmt.Errorf("missing type attribute on provider[%d]", i)
		}
		typ, ok := t.(string)
		if !ok {
			return nil, fmt.Errorf("Provider[%d]: 'type' value is not a string!", i)
		}
		builder, ok := ProviderConfigBuilderFromType[typ]
		if !ok {
			return nil, fmt.Errorf("Invalid type attribute (%s) on provider[%d]\n", t, i)
		}
		providerConfig := builder()
		data, err := yaml.Marshal(conf.Providers[i])
		if err != nil {
			return nil, err
		}
		err = yaml.UnmarshalStrict(data, providerConfig)
		if err != nil {
			return nil, err
		}
		name := providerConfig.GetName()
		if _, ok := this.prividerByName[name]; ok {
			return nil, fmt.Errorf("two providers are defined with the same name: '%s'", name)
		}
		if providerConfig.IsEnabled() {
			prvd, err := providerConfig.Open(i, conf.ConfigFolder)
			if err != nil {
				return nil, err
			}
			pcLog.Info("Setup provider", "provider", prvd.GetName())
			this.providers = append(this.providers, prvd)
			this.prividerByName[name] = prvd
		}
	}
	return &this, nil
}

func (this providerChain) String() string {
	s := ""
	sep := ""
	for _, p := range this.providers {
		s = s + sep + p.GetName()
		sep = "->"
	}
	return s
}

func (this *providerChain) Login(login, password string) (common.User, bool, string, error) {
	passwordStatus := common.Unchecked
	user := common.User{
		Username: login,
		Uid:      "",
		Groups:   []string{},
	}
	authenticator := ""
	for _, prvd := range this.providers {
		userStatus, err := prvd.GetUserStatus(login, password, passwordStatus == common.Unchecked)
		if err != nil {
			if prvd.IsCritical() {
				pcLog.Error(err, "FAIL; Provider is critical", "provider", prvd.GetName())
				return common.User{}, false, prvd.GetName(), err
			} else {
				pcLog.Error(err, "Will continue (Provider is not critical)", "provider", prvd.GetName())
				continue
			}
		}
		pcLog.Info("", "provider", prvd.GetName(), "found", userStatus.Found, "passwordStatus", userStatus.PasswordStatus, "uid", userStatus.Uid, "group", userStatus.Groups)
		if userStatus.Found {
			if userStatus.PasswordStatus == common.Wrong {
				// No need to go further. Return an empty user to avoid providing partial info
				return common.User{}, false, prvd.GetName(), nil
			}
			if userStatus.PasswordStatus == common.Checked {
				passwordStatus = common.Checked
				// The provider who validate the password is the authority for Uid
				user.Uid = userStatus.Uid
				authenticator = prvd.GetName()
			}
			user.Groups = append(user.Groups, userStatus.Groups...)
		}
	}
	if passwordStatus == common.Checked {
		user.Groups = unique(user.Groups)
		sort.Strings(user.Groups)
		return user, true, authenticator, nil
	} else {
		return common.User{}, false, authenticator, nil
	}
}

func (this *providerChain) DescribeUser(login string) (bool, common.UserDescribeResponse) {
	result := common.UserDescribeResponse{
		UserStatuses: []common.UserStatus{},
		Authority:    "",
		User: common.User{
			Username: login,
			Uid:      "",
			Groups:   []string{},
		},
	}
	found := false
	for _, prvd := range this.providers {
		userStatus, err := prvd.GetUserStatus(login, "", false)
		if err != nil {
			userStatus = common.UserStatus{
				ProviderName:   prvd.GetName(),
				PasswordStatus: common.Unchecked,
				Messages:       []string{fmt.Sprintf("Provider failure. Check server logs")},
			}
			pcLog.Error(err, "", "provider", prvd.GetName())
		} else {
			pcLog.V(1).Info("", "user", login, "provider", prvd.GetName(), "found", userStatus.Found, "passwordSatus", userStatus.PasswordStatus, "uid", userStatus.Uid, "group", userStatus.Groups, "messages", userStatus.Messages)
			if userStatus.Found {
				if result.Authority == "" && userStatus.Authority {
					result.Authority = userStatus.ProviderName
					result.User.Uid = userStatus.Uid
				}
				found = true
				result.User.Groups = append(result.User.Groups, userStatus.Groups...)
			}
		}
		result.UserStatuses = append(result.UserStatuses, userStatus)
	}
	result.User.Groups = unique(result.User.Groups)
	sort.Strings(result.User.Groups)
	return found, result
}

func (this *providerChain) GetNamespace(providerName string) (namespace string, err error) {
	if providerName == "_" {
		prvds := make([]string, 0, 10)
		ns := ""
		for _, prvd := range this.providers {
			if prvd.GetType() == "crd" {
				prvds = append(prvds, prvd.GetName())
				crdProvider := (prvd).(*crd.CrdProvider)
				ns = crdProvider.Namespace
			}
		}
		if len(prvds) > 1 {
			return "", fmt.Errorf("There is more than one provider of type 'crd'. A provider name must be given within %v!", prvds)
		} else if len(prvds) == 0 {
			return "", fmt.Errorf("There is no 'crd' provider in this configuration!")
		} else { // len(prvds) == 1
			return ns, nil
		}
	} else {
		prvd, ok := this.prividerByName[providerName]
		if ok {
			crdProvider, ok := (prvd).(*crd.CrdProvider)
			if !ok {
				return "", fmt.Errorf("Provider '%s' is not of type 'crd'.", providerName)
			}
			return crdProvider.Namespace, nil
		} else {
			return "", fmt.Errorf("There is no provider named '%s' defined in this configuration", providerName)
		}
	}
}

func (this *providerChain) GetProvider(providerName string) (providers.Provider, error) {
	p, ok := this.prividerByName[providerName]
	if !ok {
		return nil, fmt.Errorf("Provider '%s' does not exists in this configuration.", providerName)
	}
	return p, nil
}

func unique(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, exists := keys[entry]; !exists {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
