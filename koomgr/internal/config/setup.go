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

package config

import (
	"fmt"
	"github.com/koobind/koobind/koomgr/internal/utils"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func loadConfig(fileName string, config *Config) error {
	configFile, err := filepath.Abs(fileName)
	if err != nil {
		return err
	}
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	if err = decoder.Decode(&config); err != nil {
		return err
	}
	// All file path should be relative to the config file location. So take note of its absolute path
	config.ConfigFolder = filepath.Dir(configFile)
	return nil
}

func Setup() {
	// Allow overriding of some config variable. Mostly used in development stage
	var configFile string
	var logLevel int
	var logMode string
	var webhookHost string
	var webhookPort int
	var webhookCertDir string
	var authHost string
	var authPort int
	var authCertDir string
	var authNoSsl bool
	var inactivityTimeout string
	var sessionMaxTTL string
	var clientTokenTTL string
	var tokenStorage string
	var namespace string
	var tokenNamespace string
	var lastHitStep int

	pflag.StringVar(&configFile, "config", "config.yml", "Configuration file")
	pflag.IntVar(&logLevel, "logLevel", 0, "Log level (0:INFO; 1:DEBUG, 2:MoreDebug...)")
	pflag.StringVar(&logMode, "logMode", "json", "Log mode: 'dev' or 'json'")
	pflag.StringVar(&webhookHost, "webhookHost", "", "Webhook server bind address (Default: All)")
	pflag.IntVar(&webhookPort, "webhookPort", 8443, "Webhook server bind port")
	pflag.StringVar(&webhookCertDir, "webhookCertDir", "/tmp/cert/webhook-server", "Path to the webhook server certificate folder")
	pflag.StringVar(&authHost, "authHost", "", "Auth server bind address (Default: All)")
	pflag.IntVar(&authPort, "authPort", 8444, "Auth server bind port")
	pflag.StringVar(&authCertDir, "authCertDir", "/tmp/certs/auth-server", "Path to the auth server certificate folder")
	pflag.BoolVar(&authNoSsl, "authNoSsl", false, "Set the auth server in plain text. (http://). UNSECURE")
	pflag.StringVar(&inactivityTimeout, "inactivityTimeout", "30m", "Session inactivity time out")
	pflag.StringVar(&sessionMaxTTL, "sessionMaxTTL", "24h", "Session max TTL")
	pflag.StringVar(&clientTokenTTL, "clientTokenTTL", "30s", "Client local token TTL")
	pflag.StringVar(&tokenStorage, "tokenStorage", "crd", "Tokens storage mode: 'memory' or 'crd'")
	pflag.StringVar(&namespace, "namespace", "", "Default namespace for tokenNamespace and CRD")
	pflag.StringVar(&tokenNamespace, "tokenNamespace", "", "Tokens storage namespace when tokenStorage==crd")
	pflag.IntVar(&lastHitStep, "lastHitStep", 3, "Delay to store lastHit in CRD, when tokenStorage==crd. In % of inactivityTimeout")
	pflag.CommandLine.SortFlags = false
	pflag.Parse()

	err := loadConfig(configFile, &Conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Unable to load config file: %v\n", err)
		os.Exit(2)
	}
	adjustConfigInt(pflag.CommandLine, &Conf.LogLevel, "logLevel")
	adjustConfigString(pflag.CommandLine, &Conf.LogMode, "logMode")
	adjustConfigString(pflag.CommandLine, &Conf.WebhookServer.Host, "webhookHost")
	adjustConfigInt(pflag.CommandLine, &Conf.WebhookServer.Port, "webhookPort")
	adjustConfigString(pflag.CommandLine, &Conf.WebhookServer.CertDir, "webhookCertDir")
	adjustConfigString(pflag.CommandLine, &Conf.AuthServer.Host, "authHost")
	adjustConfigInt(pflag.CommandLine, &Conf.AuthServer.Port, "authPort")
	adjustConfigString(pflag.CommandLine, &Conf.AuthServer.CertDir, "authCertDir")
	adjustConfigBool(pflag.CommandLine, &Conf.AuthServer.NoSsl, "authNoSsl")
	adjustConfigDuration(pflag.CommandLine, &Conf.InactivityTimeout, "inactivityTimeout")
	adjustConfigDuration(pflag.CommandLine, &Conf.SessionMaxTTL, "sessionMaxTTL")
	adjustConfigDuration(pflag.CommandLine, &Conf.ClientTokenTTL, "clientTokenTTL")
	adjustConfigString(pflag.CommandLine, &Conf.TokenStorage, "tokenStorage")
	adjustConfigString(pflag.CommandLine, &Conf.Namespace, "namespace")
	adjustConfigString(pflag.CommandLine, &Conf.TokenNamespace, "tokenNamespace")
	adjustConfigInt(pflag.CommandLine, &Conf.LastHitStep, "lastHitStep")

	AdjustPath(Conf.ConfigFolder, &Conf.WebhookServer.CertDir)
	AdjustPath(Conf.ConfigFolder, &Conf.AuthServer.CertDir)
	if Conf.Providers == nil || len(Conf.Providers) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Missing Providers list in config file\n")
		os.Exit(2)
	}
	// CertDir, CertName and KeyName will be checked by lower layer
	if Conf.LogMode != "dev" && Conf.LogMode != "json" {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Invalid logMode value: %s. Must be one of 'dev' or 'json'\n", Conf.LogMode)
		os.Exit(2)
	}
	if Conf.LogMode == "json" && Conf.LogLevel > 1 {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: logLevel can't be greater than one when logMode is 'json'.\n")
		os.Exit(2)
	}
	if Conf.TokenStorage != "memory" && Conf.TokenStorage != "crd" {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Invalid tokenStorage value: %s. Must be one of 'memory' or 'crd'\n", Conf.LogMode)
		os.Exit(2)
	}
	if Conf.TokenNamespace == "" {
		if Conf.Namespace == "" {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: One of 'tokenNamespace' or 'namespace' parameter must be provided\n")
			os.Exit(2)
		}
		Conf.TokenNamespace = Conf.Namespace
	}
	Conf.CrdNamespaces = utils.NewStringSet()
}

func AdjustPath(baseFolder string, path *string) {
	if *path != "" {
		if !filepath.IsAbs(*path) {
			*path = filepath.Join(baseFolder, *path)
		}
		*path = filepath.Clean(*path)
	}
}

// For all adjustConfigXxx(), we:
// - panic when error is internal
// - Display a message and exit(2) when error is from usage

func adjustConfigString(flagSet *pflag.FlagSet, inConfig *string, param string) {
	if pflag.Lookup(param).Changed {
		var err error
		if *inConfig, err = flagSet.GetString(param); err != nil {
			panic(err)
		}
	} else if *inConfig == "" {
		*inConfig = flagSet.Lookup(param).DefValue
	}
}

func adjustConfigInt(flagSet *pflag.FlagSet, inConfig *int, param string) {
	var err error
	if flagSet.Lookup(param).Changed {
		if *inConfig, err = flagSet.GetInt(param); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\nInvalid value for parameter %s\n", param)
			os.Exit(2)
		}
	} else if *inConfig == 0 {
		if *inConfig, err = strconv.Atoi(flagSet.Lookup(param).DefValue); err != nil {
			panic(err)
		}
	}
}

func adjustConfigBool(flagSet *pflag.FlagSet, inConfig **bool, param string) {
	var err error
	var ljson bool
	if flagSet.Lookup(param).Changed {
		if ljson, err = flagSet.GetBool(param); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\nInvalid value for parameter %s\n", param)
			os.Exit(2)
		}
		*inConfig = &ljson
	} else if *inConfig == nil {
		if ljson, err = strconv.ParseBool(flagSet.Lookup(param).DefValue); err != nil {
			panic(err)
		}
		*inConfig = &ljson
	}
}

func adjustConfigDuration(flagSet *pflag.FlagSet, inConfig **time.Duration, param string) {
	var err error
	var durationStr string
	var duration time.Duration
	if flagSet.Lookup(param).Changed {
		if durationStr, err = flagSet.GetString(param); err != nil {
			panic(err)
		}
		if duration, err = time.ParseDuration(durationStr); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\nInvalid %s value for a duration. Must be like 300s, 20m or 12h.\n\n", param)
			os.Exit(2)

		}
		*inConfig = &duration
	} else if *inConfig == nil {
		if duration, err = time.ParseDuration(flagSet.Lookup(param).DefValue); err != nil {
			panic(err)
		}
		*inConfig = &duration
	}
}
