package kubernetes

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/qaware/minikube-support/pkg/sh"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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

	// IsMinikube returns true if the target of this context is a minikube instance. Otherwise false.
	IsMinikube() (bool, error)
}

type contextHandler struct {
	clientSet             *kubernetes.Clientset
	dynamicClient         dynamic.Interface
	clientSetMutex        sync.Mutex
	configFile            *string
	contextName           string
	predefinedContextName *string
	clientConfig          clientcmd.ClientConfig
	restConfig            *rest.Config
	minikube              *bool
}

// NewContextHandler creates a new ContextHandler instance for the given config file and context name.
func NewContextHandler(configFile *string, contextName *string) ContextHandler {
	return &contextHandler{configFile: configFile, predefinedContextName: contextName, clientSetMutex: sync.Mutex{}}
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
	return h.contextName
}

func (h *contextHandler) GetPredefinedContextName() string {
	if h.predefinedContextName != nil {
		return *h.predefinedContextName
	}
	return ""
}

func (h *contextHandler) Kubectl(command string, args ...string) (string, error) {
	prefix := []string{command}
	if h.GetPredefinedContextName() != "" {
		prefix = append(prefix, "--context", h.GetPredefinedContextName())
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

func (h *contextHandler) IsMinikube() (bool, error) {
	h.clientSetMutex.Lock()
	defer h.clientSetMutex.Unlock()

	if h.restConfig == nil {
		e := h.openRestConfig()
		if e != nil {
			return false, e
		}
	}

	if h.minikube == nil {
		ip, e := sh.RunCmd("minikube", "ip")
		if e != nil {
			if sh.IsExitCode(e, 66) {
				// minikube vm don't exists
				state := false
				h.minikube = &state
				return false, nil
			}
			return false, e
		}
		ip = strings.Trim(ip, "\n\r\t ")
		state := strings.Contains(h.restConfig.Host, ip)
		h.minikube = &state
	}
	return *h.minikube, nil
}

// openRestConfig opens the kubernetes configuration and creates a client set that can
// be used to connect to an kubernetes cluster.
func (h *contextHandler) openRestConfig() error {
	var e error
	var config *rest.Config
	configFile := *h.configFile

	if configFile == "" {
		config, e = rest.InClusterConfig()

		// if not run in cluster try to use default from user home
		if e == rest.ErrNotInCluster {
			homeDir := homedir.HomeDir()
			configPath := filepath.Join(homeDir, ".kube", "config")
			config, e = h.loadConfig(configPath)
		}

		// Neither in cluster config nor user home config exists.
		if e != nil {
			return fmt.Errorf("can not determ config: %s", e)
		}
	} else {
		// Use config from given file name.
		config, e = h.loadConfig(configFile)
		if e != nil {
			return fmt.Errorf("can not read config from file %s: %s", configFile, e)
		}
	}

	clientSet, e := kubernetes.NewForConfig(config)
	if e != nil {
		return fmt.Errorf("unable to create clientSet: %s", e)
	}
	h.clientSet = clientSet
	h.restConfig = config

	dynamicClient, e := dynamic.NewForConfig(config)
	if e != nil {
		return fmt.Errorf("unable to create dynamic client: %s", e)
	}
	h.dynamicClient = dynamicClient
	return nil
}

// loadConfig loads the actual configuration file and sets the context name.
func (h *contextHandler) loadConfig(kubeconfigPath string) (*rest.Config, error) {
	h.clientConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: h.GetPredefinedContextName()},
	)
	e := h.findUsedContextName()
	if e != nil {
		return nil, e
	}
	return h.clientConfig.ClientConfig()
}

func (h *contextHandler) findUsedContextName() error {
	if h.predefinedContextName != nil && *h.predefinedContextName != "" {
		h.contextName = *h.predefinedContextName
		return nil
	}
	config, e := h.clientConfig.RawConfig()
	if e != nil {
		return e
	}

	h.contextName = config.CurrentContext
	return nil
}
