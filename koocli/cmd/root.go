package cmd

import (
	"fmt"
	"github.com/koobind/koobind/koocli/internal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/user"
	"path"
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-koo",
	Short: "A kubectl plugin for Kubernetes authentification",
}


// package logger
var log *logrus.Entry

var context string
var config *internal.Config


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
	log.Debugf("kubeconfig=%s", kubeconfig)
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return ""
	}
	log.Debugf("kubeconfig=%s   context:%s", kubeconfig, config.CurrentContext)
	return config.CurrentContext
}

func init() {
	var rootCaFile string
	var server string
	var logLevel string
	var logJson bool
	var kubeconfig string

	rootCmd.PersistentFlags().StringVarP(&context, "context", "", "", "Context" )
	rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "", "", "Kubeconfig file path. Used to lookup context" )
	rootCmd.PersistentFlags().StringVarP(&rootCaFile, "rootCaFile", "", "", "Cert authority for client connection" )
	rootCmd.PersistentFlags().StringVarP(&server, "server", "", "", "Authentication server" )
	rootCmd.PersistentFlags().StringVarP(&logLevel, "logLevel", "", "INFO", "Log level" )
	rootCmd.PersistentFlags().BoolVarP(&logJson, "logJson", "j", false, "logs in JSON")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		internal.ConfigureLogger(logLevel, logJson)
		log = logrus.WithFields(logrus.Fields{"package": "cmd"})

		if context == "" {
			context = lookupContextInKubeconfig(kubeconfig)
			if context == "" {
				context = "default"
			}
		}
		log.Debugf("context:%s", context)

		if rootCaFile != "" {
			if !path.IsAbs(rootCaFile) {
				cwd, err := os.Getwd()
				if err != nil {
					panic(err)
				}
				rootCaFile = path.Join(cwd, rootCaFile)
			}
		}
		if cmd != configCmd {
			config = internal.LoadConfig(context)
			if config == nil {
				if server == "" {
					_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Missing 'server' parameter on initial call\n\n")
					os.Exit(2)
				}
				if rootCaFile == "" {
					_, _ = fmt.Fprintf(os.Stderr, "\nERROR: Missing 'rootCaFile' parameter on initial call\n\n")
					os.Exit(2)
				}
				config = &internal.Config{
					Server:     server,
					RootCaFile: rootCaFile,
				}
				internal.SaveConfig(context, config)
			} else {
				dirtyConfig := false
				if server != "" && server != config.Server {
					config.Server = server
					dirtyConfig = true
				}
				if rootCaFile != "" && rootCaFile != config.RootCaFile {
					config.RootCaFile = rootCaFile
					dirtyConfig = true
				}
				if dirtyConfig {
					internal.SaveConfig(context, config)
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
	if err := rootCmd.Execute(); err != nil {
		//fmt.Println(err)
		os.Exit(2)
	}
}

