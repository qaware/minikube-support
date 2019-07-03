package certmanager

import (
	"encoding/json"
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/kubernetes"
	"github.com/chr-fritz/minikube-support/pkg/packagemanager/helm"
	"github.com/chr-fritz/minikube-support/pkg/sh"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubernetes2 "k8s.io/client-go/kubernetes"
	"net/http"
	"strings"
	"time"
)

type certManager struct {
	manager        helm.Manager
	contextHandler kubernetes.ContextHandler
	clientSet      *kubernetes2.Clientset
	namespace      string
	values         map[string]interface{}
	server         string
}

const PLUGIN_NAME = "certManager"
const issuerName = "ca-issuer"

func NewCertManager(manager helm.Manager, handler kubernetes.ContextHandler) (apis.InstallablePlugin, error) {
	clientset, e := handler.GetClientSet()
	if e != nil {
		return nil, fmt.Errorf("can not get clientset: %s", e)
	}

	return &certManager{
		manager:        manager,
		contextHandler: handler,
		server:         "https://api.github.com",
		values:         map[string]interface{}{},
		clientSet:      clientset,
		namespace:      "mks",
	}, nil
}

func (m *certManager) String() string {
	return PLUGIN_NAME
}

func (m *certManager) Install() {
	if e := m.manager.AddRepository("jetstack", "https://charts.jetstack.io"); e != nil {
		logrus.Errorf("Unable to add jetstack repository: %s", e)
		return
	}
	m.Update()
}

func (m *certManager) Update() {
	version, e := m.getLatestVersion()
	if e != nil {
		logrus.Errorf("Unable to detect latest certmanager version: %s", e)
		return
	}

	downloadUrl := "https://raw.githubusercontent.com/jetstack/cert-manager/" + version + "/deploy/manifests/00-crds.yaml"

	response, e := m.contextHandler.Kubectl("apply", "-f", downloadUrl)
	if e != nil {
		logrus.Errorf("Unable to install the certmanager crds: %s", response)
		return
	}

	if e := m.manager.UpdateRepository(); e != nil {
		logrus.Errorf("Unable to update helm repositories %s", e)
		return
	}

	m.values["ingressShim.defaultIssuerName"] = issuerName
	m.values["ingressShim.defaultIssuerKind"] = "ClusterIssuer"
	m.values["webhook.enabled"] = false

	m.manager.Install("jetstack/cert-manager", releaseName, m.namespace, m.values, true)

	var errors *multierror.Error
	errors = multierror.Append(errors, m.applyCertSecret())
	errors = multierror.Append(errors, m.applyClusterIssuer())

	if errors.Len() > 0 {
		logrus.Errorf("Can not apply the additional cert manager objects: %s", errors.Error())
	}
}

func (m *certManager) Uninstall(purge bool) {
	panic("implement me")
}

func (m *certManager) Phase() apis.Phase {
	return apis.CLUSTER_TOOLS_INSTALL
}

func (m *certManager) getLatestVersion() (string, error) {
	client := http.Client{Timeout: 2 * time.Second}
	resp, e := client.Get(m.server + "/repos/jetstack/cert-manager/releases/latest")
	if e != nil {
		return "", fmt.Errorf("can not get latest version for certManager: %s", e)
	}

	data := make(map[string]interface{})
	decoder := json.NewDecoder(resp.Body)

	err := decoder.Decode(&data)
	if err != nil {
		return "", err
	}

	version, ok := data["tag_name"]
	if !ok {
		return "", fmt.Errorf("version field not found")
	}
	v, ok := version.(string)
	if !ok {
		return "", fmt.Errorf("version is not a string")
	}
	return v, nil
}

func (m *certManager) applyCertSecret() error {
	caRoot, e := sh.RunCmd("mkcert", "-CAROOT")
	if e != nil {
		return fmt.Errorf("unable to get the mkcert CA root: %s", e)
	}
	caRoot = strings.Trim(caRoot, "\r\n \t")

	crt, e := ioutil.ReadFile(caRoot + "/rootCA.pem")
	if e != nil {
		return fmt.Errorf("unable to read the mkcert RootCA certificate: %s", e)
	}
	key, e := ioutil.ReadFile(caRoot + "/rootCA-key.pem")
	if e != nil {
		return fmt.Errorf("unable to read the mkcert RootCA key: %s", e)
	}

	secretInterface := m.clientSet.CoreV1().Secrets(m.namespace)

	secret := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Namespace: m.namespace, Name: issuerName},
		Type:       v1.SecretTypeTLS,
		Data: map[string][]byte{
			v1.TLSCertKey:       crt,
			v1.TLSPrivateKeyKey: key,
		},
	}

	_, e = secretInterface.Get(issuerName, metav1.GetOptions{})
	if e != nil {
		_, e = secretInterface.Create(secret)
	} else {
		_, e = secretInterface.Update(secret)
	}

	if e != nil {
		return fmt.Errorf("applying the secret failed: %s", e)
	}
	return nil
}

func (m *certManager) applyClusterIssuer() error {
	client, e := m.contextHandler.GetDynamicClient()
	if e != nil {
		return e
	}
	groupVersion := schema.GroupVersion{Group: "certmanager.k8s.io", Version: "v1alpha1"}

	issuer := &unstructured.Unstructured{}
	issuer.SetUnstructuredContent(map[string]interface{}{"spec": map[string]interface{}{"ca": map[string]interface{}{"secretName": issuerName}}})
	issuer.SetGroupVersionKind(groupVersion.WithKind("ClusterIssuer"))
	issuer.SetName(issuerName)

	clusterIssuerInterface := client.Resource(groupVersion.WithResource("clusterissuers"))
	_, e = clusterIssuerInterface.Get(issuerName, metav1.GetOptions{})
	if e != nil {
		_, e = clusterIssuerInterface.Create(issuer, metav1.CreateOptions{})
	} else {
		_, e = clusterIssuerInterface.Update(issuer, metav1.UpdateOptions{})
	}
	if e != nil {
		return fmt.Errorf("applying the cluster issuer failed: %s", e)
	}
	return nil
}

// k8s.io/apimachinery/pkg/api/errors.StatusError -> ErrStatus -> StatusReason=NotFound
