package config

import "sigs.k8s.io/controller-runtime/pkg/client"

// This file host ALL global variables.

var (
	// THE GLOBAL CONFIGURATION SINGLETON
	Conf = Config{}

	KubeClient client.Client
)
