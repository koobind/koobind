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
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/providers/crd"
	"github.com/koobind/koobind/koomgr/internal/providers/ldap"
	"github.com/koobind/koobind/koomgr/internal/providers/static"
	"gopkg.in/yaml.v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sort"
)

var _ providers.ProviderChain = &providerChain{}

type providerChain struct {
	providers      []providers.Provider
	providerByName map[string]providers.Provider
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
		providerByName: make(map[string]providers.Provider),
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
		if _, ok := this.providerByName[name]; ok {
			return nil, fmt.Errorf("two providers are defined with the same name: '%s'", name)
		}
		if providerConfig.IsEnabled() {
			prvd, err := providerConfig.Open(i, conf.ConfigFolder)
			if err != nil {
				return nil, err
			}
			pcLog.Info("Setup provider", "provider", prvd.GetName())
			this.providers = append(this.providers, prvd)
			this.providerByName[name] = prvd
		}
	}
	return &this, nil
}

func (this *providerChain) String() string {
	s := ""
	sep := ""
	for _, p := range this.providers {
		s = s + sep + p.GetName()
		sep = "->"
	}
	return s
}

func (this *providerChain) Login(login, password string) (tokenapi.UserDesc, bool, error) {
	passwordStatus := tokenapi.PasswordStatusUnchecked
	user := tokenapi.UserDesc{
		Name:        login,
		Groups:      []string{},
		Emails:      []string{},
		CommonNames: []string{},
		Entries:     []tokenapi.UserEntry{},
	}
	for _, prvd := range this.providers {
		userEntry, err := prvd.GetUserStatus(login, password, passwordStatus == tokenapi.PasswordStatusUnchecked)
		user.Entries = append(user.Entries, userEntry)
		if err != nil {
			if prvd.IsCritical() {
				pcLog.Error(err, "FAIL; Provider is critical", "provider", prvd.GetName())
				return tokenapi.UserDesc{Authority: prvd.GetName()}, false, err
			} else {
				pcLog.Error(err, "Will continue (Provider is not critical)", "provider", prvd.GetName())
				continue
			}
		}
		pcLog.Info("", "provider", prvd.GetName(), "found", userEntry.Found, "passwordStatus", userEntry.PasswordStatus, "uid", userEntry.Uid, "group", userEntry.Groups)
		if userEntry.Found {
			if userEntry.PasswordStatus == tokenapi.PasswordStatusWrong {
				// No need to go further. Return an almost empty user to avoid providing partial info
				return tokenapi.UserDesc{Authority: prvd.GetName()}, false, nil
			}
			if userEntry.PasswordStatus == tokenapi.PasswordStatusChecked {
				passwordStatus = tokenapi.PasswordStatusChecked
				// The provider who validate the password is the authority for Uid
				user.Uid = userEntry.Uid
				user.Authority = prvd.GetName()
			}
			user.Groups = append(user.Groups, userEntry.Groups...)
			if userEntry.Email != "" {
				user.Emails = append(user.Emails, userEntry.Email)
			}
			user.CommonNames = append(user.CommonNames, userEntry.CommonName)
		}
	}
	if passwordStatus != tokenapi.PasswordStatusChecked {
		return tokenapi.UserDesc{}, false, nil
	}
	user.Groups = dedupAndSort(user.Groups)
	user.Emails = dedupAndSort(user.Emails)
	user.CommonNames = dedupAndSort(user.CommonNames)
	return user, true, nil
}

func (this *providerChain) DescribeUser(login string) (bool, tokenapi.UserDesc) {
	user := tokenapi.UserDesc{
		Name:        login,
		Groups:      []string{},
		Emails:      []string{},
		CommonNames: []string{},
		Entries:     []tokenapi.UserEntry{},
	}
	found := false
	for _, prvd := range this.providers {
		userEntry, err := prvd.GetUserStatus(login, "", false)
		if err != nil {
			// Build a substituted userEntry
			userEntry = tokenapi.UserEntry{
				ProviderName:   prvd.GetName(),
				PasswordStatus: tokenapi.PasswordStatusUnchecked,
				Messages:       []string{fmt.Sprintf("Provider failure. Check server logs")},
			}
			pcLog.Error(err, "", "provider", prvd.GetName())
		} else {
			pcLog.V(1).Info("", "user", login, "provider", prvd.GetName(), "found", userEntry.Found, "passwordSatus", userEntry.PasswordStatus, "uid", userEntry.Uid, "group", userEntry.Groups, "messages", userEntry.Messages)
			if userEntry.Found {
				if user.Authority == "" && userEntry.Authority {
					user.Authority = userEntry.ProviderName
					user.Uid = userEntry.Uid
				}
				found = true
				user.Groups = append(user.Groups, userEntry.Groups...)
				if userEntry.Email != "" {
					user.Emails = append(user.Emails, userEntry.Email)
				}
				user.CommonNames = append(user.CommonNames, userEntry.CommonName)
			}
		}
		user.Entries = append(user.Entries, userEntry)
	}
	user.Groups = dedupAndSort(user.Groups)
	user.Emails = dedupAndSort(user.Emails)
	user.CommonNames = dedupAndSort(user.CommonNames)
	return found, user
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
		prvd, ok := this.providerByName[providerName]
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
	p, ok := this.providerByName[providerName]
	if !ok {
		return nil, fmt.Errorf("Provider '%s' does not exists in this configuration.", providerName)
	}
	return p, nil
}

func dedupAndSort(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, exists := keys[entry]; !exists {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	sort.Strings(list)
	return list
}
