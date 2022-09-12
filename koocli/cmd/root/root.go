/*
Copyright (C) 2020 Serge ALEXANDRE

# This file is part of koobind project

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
package root

import (
	"fmt"
	. "github.com/koobind/koobind/koocli/cmd/common"
	koocontext "github.com/koobind/koobind/koocli/cmd/context"
	"github.com/koobind/koobind/koocli/cmd/group"
	"github.com/koobind/koobind/koocli/cmd/groupbinding"
	"github.com/koobind/koobind/koocli/cmd/misc"
	"github.com/koobind/koobind/koocli/cmd/token"
	koouser "github.com/koobind/koobind/koocli/cmd/user"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/user"
	"path"
)

var RootCmd = &cobra.Command{
	Use:   "kubectl-koo",
	Short: "A kubectl plugin for Kubernetes authentification",
}

func lookupContextInKubeconfig(kubeconfig string) string {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	if kubeconfig == "" {
		usr, err := user.Current()
		if err == nil {
			kubeconfig = path.Join(usr.HomeDir, ".kube/config")
		}
	}
	Log.Debugf("kubeconfig=%s", kubeconfig)
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return ""
	}
	Log.Debugf("kubeconfig=%s   Context:%s", kubeconfig, config.CurrentContext)
	return config.CurrentContext
}

func init() {
	var rootCaFile string
	var server string
	var logLevel string
	var logJson bool
	var kubeconfig string

	// We must declare child in parent.
	// Performing RootCmd.AddCommand(...) in the child init() function will not works as there is chance the child package will not be loaded, as not imported by any package.
	RootCmd.AddCommand(misc.AuthCmd)
	RootCmd.AddCommand(misc.HashCmd)
	RootCmd.AddCommand(misc.LoginCmd)
	RootCmd.AddCommand(misc.LogoutCmd)
	RootCmd.AddCommand(misc.WhoamiCmd)
	RootCmd.AddCommand(misc.VersionCmd)
	RootCmd.AddCommand(misc.PasswordCmd)

	RootCmd.AddCommand(token.TokenCmd)
	RootCmd.AddCommand(koocontext.ContextCmd)
	RootCmd.AddCommand(koouser.UserCmd)
	RootCmd.AddCommand(group.GroupCmd)
	RootCmd.AddCommand(groupbinding.GroupBindingCmd)

	RootCmd.PersistentFlags().StringVar(&Context, "Context", "", "Context")
	RootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Kubeconfig file path. Used to lookup Context")
	RootCmd.PersistentFlags().StringVar(&rootCaFile, "rootCaFile", "", "Cert authority for client connection")
	RootCmd.PersistentFlags().StringVar(&server, "server", "", "Authentication server")
	RootCmd.PersistentFlags().StringVar(&logLevel, "logLevel", "INFO", "Log level")
	RootCmd.PersistentFlags().BoolVar(&logJson, "logJson", false, "logs in JSON")

	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		internal.ConfigureLogger(logLevel, logJson)
		Log = logrus.WithFields(logrus.Fields{"package": "cmd"})

		if Context == "" {
			Context = lookupContextInKubeconfig(kubeconfig)
			if Context == "" {
				Context = "default"
			}
		}
		Log.Debugf("Context:%s", Context)

		if rootCaFile != "" {
			if !path.IsAbs(rootCaFile) {
				cwd, err := os.Getwd()
				if err != nil {
					panic(err)
				}
				rootCaFile = path.Join(cwd, rootCaFile)
			}
		}
		if cmd != koocontext.ContextCmd {
			Config = internal.LoadConfig(Context)
			if Config == nil {
				if server == "" {
					_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Missing 'server' parameter on initial call\n\n")
					os.Exit(2)
				}
				if rootCaFile == "" {
					_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Missing 'rootCaFile' parameter on initial call\n\n")
					os.Exit(2)
				}
				Config = &internal.Config{
					Server:     server,
					RootCaFile: rootCaFile,
				}
				internal.SaveConfig(Context, Config)
			} else {
				dirtyConfig := false
				if server != "" && server != Config.Server {
					Config.Server = server
					dirtyConfig = true
				}
				if rootCaFile != "" && rootCaFile != Config.RootCaFile {
					Config.RootCaFile = rootCaFile
					dirtyConfig = true
				}
				if dirtyConfig {
					internal.SaveConfig(Context, Config)
				}
			}
		}
	}
}

var debug = true

func Execute() {
	defer func() {
		if !debug {
			if r := recover(); r != nil {
				fmt.Printf("ERROR:%v\n", r)
				os.Exit(1)
			}
		}
	}()
	if err := RootCmd.Execute(); err != nil {
		//fmt.Println(err)
		os.Exit(2)
	}
}
