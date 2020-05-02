package config

import (
	"fmt"
	"time"
)

// THE GLOBAL CONFIGURATION SINGLETON
var Conf = Config{}

type Server struct {
	Host    string `yaml:"host"`    // Host is the address that the server will listen on. Defaults to "" - all addresses.
	Port    int    `yaml:"port"`    // Port is the port number that the server will serve. It will be defaulted to 443 if unspecified.
	CertDir string `yaml:"certDir"` // CertDir is the directory that contains the server key and certificate.
}

type Config struct {
	ConfigFolder      string          // This is not in the file, but set on reading. Used to adjust file path
	WebhookServer     Server          `yaml:"webhookServer"`     // The server for the mutating/validating and authentication webhook. Called only by API Server
	AuthServer        Server          `yaml:"authServer"`        // The server for authentication. To be exposed externally. Called by koocli
	LogLevel          int             `yaml:"logLevel"`          // Log level. 0: Info, 1: Debug, 2: Trace, ... Default is 0.
	LogMode           string          `yaml:"logMode"`           // Log output format: 'dev' or 'json'
	AdminGroup        string          `yaml:"adminGroup"`        // Only user belonging to this group will be able to access admin interface
	InactivityTimeout *time.Duration  `yaml:"inactivityTimeout"` // After this period without token validation, the session expire
	SessionMaxTTL     *time.Duration  `yaml:"sessionMaxTTL"`     // After this period, the session expire, in all case.
	ClientTokenTTL    *time.Duration  `yaml:"clientTokenTTL"`    // This is intended for the client (koocli), for token caching
	TokenStorage      string          `yaml:"tokenStorage"`      // 'memory' or 'crd'
	TokenNamespace    string          `yaml:"tokenNamespace"`    // When tokenStorage==crd, the namespace to store them. Default to 'koo-system'
	LastHitStep       int             `yaml:"lastHitStep"`       // When tokenStorage==crd, the max difference between reality and what is stored in API Server. In per mille of InactivityTimeout. Aim is to avoid API servr overloading
	Providers         []interface{}   `yaml:"providers"`         // The ordered list of ID providers
	Namespaces        map[string]bool // Not in the file, but used by validating webhook
}

type BaseProviderConfig struct {
	Name                string `yaml:"name"`
	Type                string `yaml:"type"`
	Enabled             *bool  `yaml:"enabled"`             // Allow to disable a provider
	CredentialAuthority *bool  `yaml:"credentialAuthority"` // Is this ldap is authority for password checking
	GroupAuthority      *bool  `yaml:"groupAuthority"`      // Group will be fetched. Default true
	Critical            *bool  `yaml:"critical"`            // If true (default), a failure on this provider will leads 'invalid login'. Even if another provider grants access
	GroupPattern        string `yaml:"groupPattern"`        // Group pattern. Default "%s"
	UidOffet            int    `yaml:"uidOffset"`           // Will be added to the returned offset. Default to 0
}

func (this *BaseProviderConfig) InitBase(idx int) error {
	// Type already checked by the builder
	// Test required fields
	if this.Name == "" {
		return fmt.Errorf("Missing required provider[%d].name", idx)
	}
	// Set default values
	yes := true
	if this.CredentialAuthority == nil {
		this.CredentialAuthority = &yes
	}
	if this.GroupAuthority == nil {
		this.GroupAuthority = &yes
	}
	if this.Critical == nil {
		this.Critical = &yes
	}
	if this.GroupPattern == "" {
		this.GroupPattern = "%s"
	}
	return nil
}

func (this *BaseProviderConfig) GetName() string {
	return this.Name
}

// Default setting (initBase) is not performed when this is called
func (this *BaseProviderConfig) IsEnabled() bool {
	return this.Enabled == nil || *this.Enabled
}
