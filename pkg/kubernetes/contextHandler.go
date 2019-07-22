package kubernetes

import (
	"fmt"
	"github.com/qaware/minikube-support/pkg/sh"
	"k8s.io/client-go/dynamic"
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
	GetClientSet() (kubernetes.Interface, error)

	// GetDynamicClient gets the kubernetes dynamic client to access unknown custom resources.
	GetDynamicClient() (dynamic.Interface, error)

	// GetConfigFile gets the path to the configuration file.
	GetConfigFile() string

	// GetContextName get the name of the context to use.
	GetContextName() string

	// Kubectl executes the given command with the given arguments in the configured cluster.
	Kubectl(command string, args ...string) (string, error)
}

type contextHandler struct {
	clientSet      *kubernetes.Clientset
	dynamicClient  dynamic.Interface
	clientSetMutex sync.Mutex
	configFile     *string
	contextName    *string
}

// NewContextHandler creates a new ContextHandler instance for the given config file and context name.
func NewContextHandler(configFile *string, contextName *string) ContextHandler {
	return &contextHandler{configFile: configFile, contextName: contextName, clientSetMutex: sync.Mutex{}}
}

func (h *contextHandler) GetClientSet() (kubernetes.Interface, error) {
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

func (h *contextHandler) GetDynamicClient() (dynamic.Interface, error) {
	h.clientSetMutex.Lock()
	defer h.clientSetMutex.Unlock()

	if h.dynamicClient == nil {
		e := h.openRestConfig()
		if e != nil {
			return nil, e
		}
	}

	return h.dynamicClient, nil
}

func (h *contextHandler) GetConfigFile() string {
	if h.configFile != nil {
		return *h.configFile
	}
	return ""
}

func (h *contextHandler) GetContextName() string {
	if h.contextName != nil {
		return *h.contextName
	}
	return ""
}

func (h *contextHandler) Kubectl(command string, args ...string) (string, error) {
	prefix := append([]string{command})
	if h.GetContextName() != "" {
		prefix = append(prefix, "--context", h.GetContextName())
	}
	if h.GetConfigFile() != "" {
		prefix = append(prefix, "--kubeconfig", h.GetConfigFile())
	}

	args = append(prefix, args...)

	output, e := sh.RunCmd("kubectl", args...)
	if e != nil {
		return output, fmt.Errorf("run kubectl %s was not successful: (%s) %s", command, e, output)
	}
	return output, nil
}

// openRestConfig opens the kubernetes configuration and creates a client set that can
// be used to connect to an kubernetes cluster.
func (h *contextHandler) openRestConfig() error {
	var e error
	var config *rest.Config
	configFile := *h.configFile
	contextName := *h.contextName

	if configFile == "" {
		config, e = rest.InClusterConfig()

		// if not run in cluster try to use default from user home
		if e == rest.ErrNotInCluster {
			homeDir := homedir.HomeDir()
			configPath := filepath.Join(homeDir, ".kube", "config")
			config, e = loadConfig(configPath, contextName)
		}

		// Neither in cluster config nor user home config exists.
		if e != nil {
			return fmt.Errorf("can not determ config: %s", e)
		}
	} else {
		// Use config from given file name.
		config, e = loadConfig(configFile, contextName)
		if e != nil {
			return fmt.Errorf("can not read config from file %s: %s", configFile, e)
		}
	}

	clientSet, e := kubernetes.NewForConfig(config)
	if e != nil {
		return fmt.Errorf("unable to create clientSet: %s", e)
	}
	h.clientSet = clientSet

	dynamicClient, e := dynamic.NewForConfig(config)
	if e != nil {
		return fmt.Errorf("unable to create dynamic client: %s", e)
	}
	h.dynamicClient = dynamicClient
	return nil
}

// loadConfig loads the actual configuration file and sets the context name.
func loadConfig(kubeconfigPath string, contextName string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: contextName},
	).ClientConfig()
}
