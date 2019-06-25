package kubernetes

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"sync"
)

// ContextHandler is a small interface to encapsulate everything related with
// retrieving the kubernetes configuration and context.
type ContextHandler interface {
	// GetClientSet gets the kubernetes client set for the defined config file and context.
	GetClientSet() (*kubernetes.Clientset, error)

	// GetConfigFile gets the path to the configuration file.
	GetConfigFile() string

	// GetContextName get the name of the context to use.
	GetContextName() string
}

type contextHandler struct {
	clientSet      *kubernetes.Clientset
	clientSetMutex sync.Mutex
	configFile     string
	contextName    string
}

// NewContextHandler creates a new ContextHandler instance for the given config file and context name.
func NewContextHandler(configFile string, contextName string) ContextHandler {
	return &contextHandler{configFile: configFile, contextName: contextName, clientSetMutex: sync.Mutex{}}
}

func (h *contextHandler) GetClientSet() (*kubernetes.Clientset, error) {
	h.clientSetMutex.Lock()
	defer h.clientSetMutex.Unlock()

	if h.clientSet == nil {
		e := h.openRestConfig()
		if e != nil {
			return nil, e
		}
	}

	return h.clientSet, nil
}

func (h *contextHandler) GetConfigFile() string {
	return h.configFile
}

func (h *contextHandler) GetContextName() string {
	return h.contextName
}

// openRestConfig opens the kubernetes configuration and creates a client set that can
// be used to connect to an kubernetes cluster.
func (h *contextHandler) openRestConfig() error {
	var e error
	var config *rest.Config
	if h.configFile == "" {
		config, e = rest.InClusterConfig()

		// if not run in cluster try to use default from user home
		if e == rest.ErrNotInCluster {
			homeDir := homedir.HomeDir()
			configPath := filepath.Join(homeDir, ".kube", "config")
			config, e = loadConfig(configPath, h.contextName)
		}

		// Neither in cluster config nor user home config exists.
		if e != nil {
			return fmt.Errorf("can not determ config: %s", e)
		}
	} else {
		// Use config from given file name.
		config, e = loadConfig(h.configFile, h.contextName)
		if e != nil {
			return fmt.Errorf("can not read config from file %s: %s", h.configFile, e)
		}
	}

	clientSet, e := kubernetes.NewForConfig(config)
	if e != nil {
		return fmt.Errorf("unable to create clientSet: %s", e)
	}
	h.clientSet = clientSet
	return nil
}

// loadConfig loads the actual configuration file and sets the context name.
func loadConfig(kubeconfigPath string, contextName string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: contextName},
	).ClientConfig()
}
