package ldap

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/koobind/koobind/koomgr/internal/config"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"io/ioutil"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NB: These values are strongly inspired from dex configuration (https://github.com/dexidp/dex)
type LdapProviderConfig struct {
	config.BaseProviderConfig `yaml:",inline"`

	// The host and port of the LDAP server.
	// If port isn't supplied, it will be guessed based on the TLS configuration. 389 or 636.
	Host string `yaml:"host"`
	Port string `yaml:"port"`

	// Required if LDAP host does not use TLS.
	InsecureNoSSL bool `yaml:"insecureNoSSL"`

	// Don't verify the CA.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`

	// Connect to the insecure port then issue a StartTLS command to negotiate a
	// secure connection. If unsupplied secure connections will use the LDAPS
	// protocol.
	StartTLS bool `yaml:"startTLS"`

	// Path to a trusted root certificate file.
	RootCA string `yaml:"rootCA"`
	// Base64 encoded PEM data containing root CAs.
	RootCAData []byte `yaml:"rootCAData"`
	// Path to a client cert file
	ClientCert string `yaml:"clientCert"`
	// Path to a client private key file
	ClientKey string `yaml:"clientKey"`

	// BindDN and BindPW for an application service account. The connector uses these
	// credentials to search for users and groups.
	BindDN string `yaml:"bindDN"`
	BindPW string `yaml:"bindPW"`

	UserSearch struct {
		// BaseDN to start the search from. For example "cn=users,dc=example,dc=com"
		BaseDN string `yaml:"baseDN"`

		// Optional filter to apply when searching the directory. For example "(objectClass=person)"
		Filter string `yaml:"filter"`

		// Attribute to match against the login. This will be translated and combined
		// with the other filter as "(<loginAttr>=<login>)".
		LoginAttr string `yaml:"loginAttr"`

		// Can either be:
		// * "sub" - search the whole sub tree
		// * "one" - only search one level
		Scope string `yaml:"scope"`

		// The attribute providing the numerical user ID
		NumericalIdAttr string `yaml:"numericalIdAttr"`

		// The attribute providing the user's email
		EmailAttr string `yaml:"emailAttr"`

		// The attribute providing the user's common name
		CnAttr string `yaml:"cnAttr"`
	} `yaml:"userSearch"`

	// Group search configuration.
	GroupSearch struct {
		// BaseDN to start the search from. For example "cn=groups,dc=example,dc=com"
		BaseDN string `yaml:"baseDN"`

		// Optional filter to apply when searching the directory. For example "(objectClass=posixGroup)"
		Filter string `yaml:"filter"`

		Scope string `yaml:"scope"` // Defaults to "sub"

		// The attribute of the group that represents its name.
		NameAttr string `yaml:"nameAttr"`

		// The filter for group/user relationship will be: (<linkGroupAttr>=<Value of LinkUserAttr for the user>)
		// If there is several value for LinkUserAttr, we will loop on.
		LinkUserAttr  string `yaml:"linkUserAttr"`
		LinkGroupAttr string `yaml:"linkGroupAttr"`
	} `yaml:"groupSearch"`
}

func (this *LdapProviderConfig) Open(idx int, configFolder string) (providers.Provider, error) {
	if err := this.InitBase(idx); err != nil {
		return nil, err
	}
	prvd := ldapProvider{
		LdapProviderConfig: this,
	}
	if prvd.Host == "" {
		return &prvd, fmt.Errorf("Missing required providers.%s.host", prvd.Name)
	}
	if prvd.UserSearch.BaseDN == "" {
		return &prvd, fmt.Errorf("Missing required providers.%s.userSearch.baseDN", prvd.Name)
	}
	if prvd.UserSearch.LoginAttr == "" {
		return &prvd, fmt.Errorf("Missing required providers.%s.userSearch.loginAttr", prvd.Name)
	}
	if *prvd.GroupAuthority {
		if prvd.GroupSearch.BaseDN == "" {
			return &prvd, fmt.Errorf("Missing required providers.%s.groupSearch.baseDN", prvd.Name)
		}
		if prvd.GroupSearch.NameAttr == "" {
			return &prvd, fmt.Errorf("Missing required providers.%s.groupSearch.nameAttr", prvd.Name)
		}
		if prvd.GroupSearch.LinkGroupAttr == "" {
			return &prvd, fmt.Errorf("Missing required providers.%s.groupSearch.linkGroupAttr", prvd.Name)
		}
		if prvd.GroupSearch.LinkUserAttr == "" {
			return &prvd, fmt.Errorf("Missing required providers.%s.groupSearch.linkUserAttr", prvd.Name)
		}
	}
	prvd.logger = ctrl.Log.WithName("ldap:" + prvd.Name)
	// Setup default value
	if prvd.Port == "" {
		if prvd.InsecureNoSSL {
			prvd.Port = "389"
		} else {
			prvd.Port = "636"
		}
	}
	prvd.hostPort = fmt.Sprintf("%s:%s", prvd.Host, prvd.Port)

	config.AdjustPath(configFolder, &prvd.RootCA)
	config.AdjustPath(configFolder, &prvd.ClientCert)
	config.AdjustPath(configFolder, &prvd.ClientKey)

	prvd.tlsConfig = &tls.Config{ServerName: prvd.Host, InsecureSkipVerify: prvd.InsecureSkipVerify}
	if prvd.RootCA != "" || len(prvd.RootCAData) != 0 {
		data := prvd.RootCAData
		if len(data) == 0 {
			var err error
			if data, err = ioutil.ReadFile(prvd.RootCA); err != nil {
				return &prvd, fmt.Errorf("Read CA file: %v", err)
			}
		}
		rootCAs := x509.NewCertPool()
		if !rootCAs.AppendCertsFromPEM(data) {
			return &prvd, fmt.Errorf("No certs found in ca file")
		}
		prvd.tlsConfig.RootCAs = rootCAs
	}

	if prvd.ClientKey != "" && prvd.ClientCert != "" {
		cert, err := tls.LoadX509KeyPair(prvd.ClientCert, prvd.ClientKey)
		if err != nil {
			return &prvd, fmt.Errorf("Load client cert failed: %v", err)
		}
		prvd.tlsConfig.Certificates = append(prvd.tlsConfig.Certificates, cert)
	}
	var ok bool
	prvd.userSearchScope, ok = parseScope(prvd.UserSearch.Scope)
	if !ok {
		return &prvd, fmt.Errorf("userSearch.Scope unknown value %q", prvd.UserSearch.Scope)
	}
	prvd.groupSearchScope, ok = parseScope(prvd.GroupSearch.Scope)
	if !ok {
		return &prvd, fmt.Errorf("groupSearch.Scope unknown value %q", prvd.GroupSearch.Scope)
	}
	return &prvd, nil
}
