package chain

import (
	"fmt"
	"github.com/golang-collections/collections/set"
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/providers/crd"
	"github.com/koobind/koobind/koomgr/internal/providers/ldap"
	"github.com/koobind/koobind/koomgr/internal/providers/static"
	"gopkg.in/yaml.v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
)

type providerChain struct {
	providers []providers.Provider
}

var pcLog = ctrl.Log.WithName("providerChain")

type providerConfig interface {
	Open(idx int, configFolder string, kubeClient client.Client) (providers.Provider, error)
	GetName() string
	IsEnabled() bool
}

var ProviderConfigBuilderFromType = map[string]func() providerConfig{
	"static": func() providerConfig { return new(static.StaticProviderConfig) },
	"ldap":   func() providerConfig { return new(ldap.LdapProviderConfig) },
	"crd":    func() providerConfig { return new(crd.CrdProviderConfig) },
}

func BuildProviderChain(conf *config.Config, kubeClient client.Client) (providers.ProviderChain, error) {
	this := providerChain{
		providers: []providers.Provider{},
	}
	providerNameSet := set.New()
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
		if providerNameSet.Has(name) {
			return nil, fmt.Errorf("two providers are defined with the same name: '%s'", name)
		}
		providerNameSet.Insert(name)
		if providerConfig.IsEnabled() {
			prvd, err := providerConfig.Open(i, conf.ConfigFolder, kubeClient)
			if err != nil {
				return nil, err
			}
			pcLog.Info("Setup provider", "provider", prvd.GetName())
			this.providers = append(this.providers, prvd)
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

func (this *providerChain) DescribeUser(login string) ([]common.UserStatus, error) {
	userStatuses := []common.UserStatus{}
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
		}
		userStatuses = append(userStatuses, userStatus)
	}
	return userStatuses, nil
}
