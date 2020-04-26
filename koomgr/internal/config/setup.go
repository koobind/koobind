package config

import (
	"fmt"
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
	// Allow overridng of some config variable. Mosty used in development stage
	var configFile string
	var logLevel int
	var host string
	var port int
	var certDir string
	var namespace string
	var inactivityTimeout string
	var sessionMaxTTL string
	var clientTokenTTL string

	pflag.StringVar(&configFile, "config", "config.yml", "Configuration file")
	pflag.IntVar(&logLevel, "logLevel", 0, "Log level (0:INFO; 1:DEBUG, 2:MoreDebug...)")
	pflag.StringVar(&host, "host", "", "Server bind address (Default: All)")
	pflag.IntVar(&port, "port", 8443, "Server bind port")
	pflag.StringVar(&certDir, "certDir", "", "Path to the server certificate folder")
	pflag.StringVar(&namespace, "namespace", "", "The namespace where koo resources (users,groups,bindings) are stored")
	pflag.StringVar(&inactivityTimeout, "inactivityTimeout", "30m", "Session inactivity time out")
	pflag.StringVar(&sessionMaxTTL, "sessionMaxTTL", "24h", "Session max TTL")
	pflag.StringVar(&clientTokenTTL, "clientTokenTTL", "30s", "Client local token TTL")
	pflag.CommandLine.SortFlags = false
	pflag.Parse()

	err := loadConfig(configFile, &Conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Unable to load config file: %v\n", err)
		os.Exit(2)
	}
	adjustConfigInt(pflag.CommandLine, &Conf.LogLevel, "logLevel")
	adjustConfigString(pflag.CommandLine, &Conf.Host, "host")
	adjustConfigInt(pflag.CommandLine, &Conf.Port, "port")
	adjustConfigString(pflag.CommandLine, &Conf.CertDir, "certDir")
	adjustConfigString(pflag.CommandLine, &Conf.Namespace, "namespace")
	adjustConfigDuration(pflag.CommandLine, &Conf.InactivityTimeout, "inactivityTimeout")
	adjustConfigDuration(pflag.CommandLine, &Conf.SessionMaxTTL, "sessionMaxTTL")
	adjustConfigDuration(pflag.CommandLine, &Conf.ClientTokenTTL, "clientTokenTTL")

	AdjustPath(Conf.ConfigFolder, &Conf.CertDir)
	if Conf.Providers == nil || len(Conf.Providers) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Missing Providers list in config file\n")
		os.Exit(2)
	}
	if Conf.Namespace == "" {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: config.namespace definition is required or --namespace parameter must be provided!\n")
		os.Exit(2)
	}
	// CertDir, CertName and KeyName will be checked by lower layer

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
