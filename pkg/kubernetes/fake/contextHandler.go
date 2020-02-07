package fake

import (
	"fmt"

	"github.com/qaware/minikube-support/pkg/testutils"
	"k8s.io/client-go/dynamic"
	dyntestclient "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
)

// ContextHandler is a simple context handler for unit tests.
type ContextHandler struct {
	ClientSet        *testclient.Clientset
	DynamicClient    *dyntestclient.FakeDynamicClient
	ConfigFile       string
	ContextName      string
	kubectlResponses []testutils.TestProcessResponse
	MiniKube         bool
}

// NewContextHandler initializes a new ContextHandler instance for unit tests.
func NewContextHandler(clientSet *testclient.Clientset, dynamicClient *dyntestclient.FakeDynamicClient) *ContextHandler {
	return &ContextHandler{
		ClientSet:        clientSet,
		DynamicClient:    dynamicClient,
		kubectlResponses: []testutils.TestProcessResponse{},
	}
}

func (f *ContextHandler) GetClientSet() (kubernetes.Interface, error) {
	if f.ClientSet == nil {
		return nil, fmt.Errorf("no client set")
	}
	return f.ClientSet, nil
}

func (f *ContextHandler) GetDynamicClient() (dynamic.Interface, error) {
	if f.DynamicClient == nil {
		return nil, fmt.Errorf("no dynamic client")
	}
	return f.DynamicClient, nil
}

func (f *ContextHandler) GetConfigFile() string {
	return f.ConfigFile
}

func (f *ContextHandler) GetContextName() string {
	return f.ContextName
}

func (f *ContextHandler) Kubectl(command string, args ...string) (string, error) {
	response := testutils.FindTestProcessResponse(f.kubectlResponses, "kubectl", append([]string{command}, args...))

	if response.ResponseStatus != 0 {
		return response.Stderr, fmt.Errorf("error response status: %v", response.ResponseStatus)
	}
	return response.Stdout, nil
}

func (f *ContextHandler) MockKubectl(command string, args []string, response string, responseStatus int) {
	f.kubectlResponses = append(f.kubectlResponses, testutils.TestProcessResponse{
		Command:        "kubectl",
		Args:           append([]string{command}, args...),
		Stdout:         response,
		ResponseStatus: responseStatus,
	})
}

func (f *ContextHandler) IsMinikube() (bool, error) {
	return f.MiniKube, nil
}
