package certmanager

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/github"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/packagemanager/helm"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type certManager struct {
	manager        helm.Manager
	contextHandler kubernetes.ContextHandler
	ghClient       github.Client
	namespace      string
	values         map[string]interface{}
	ctx            context.Context
}

const PluginName = "certManager"
const issuerName = "ca-issuer"
const releaseName = "cert-manager"

var groupVersion = schema.GroupVersion{Group: "cert-manager.io", Version: "v1"}
var helmInstallWaitPeriod = 20 * time.Second

func NewCertManager(manager helm.Manager, handler kubernetes.ContextHandler, ghClient github.Client) apis.InstallablePlugin {
	return &certManager{
		manager:        manager,
		contextHandler: handler,
		ghClient:       ghClient,
		values:         map[string]interface{}{},
		namespace:      "mks",
		ctx:            context.Background(),
	}
}

func (m *certManager) String() string {
	return PluginName
}

func (m *certManager) Install() {
	if e := m.manager.AddRepository("jetstack", "https://charts.jetstack.io"); e != nil {
		logrus.Errorf("Unable to add jetstack repository: %s", e)
		return
	}
	m.Update()
}

func (m *certManager) Update() {
	version, e := m.ghClient.GetLatestReleaseTag("jetstack", "cert-manager")
	if e != nil {
		logrus.Errorf("Unable to detect latest certmanager version: %s", e)
		return
	}

	downloadUrl := "https://github.com/jetstack/cert-manager/releases/download/" + version + "/cert-manager.crds.yaml"

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
	m.values["ingressShim.defaultIssuerGroup"] = "cert-manager.io"

	m.manager.Install("jetstack/cert-manager", releaseName, m.namespace, m.values, true)

	var err *multierror.Error
	time.Sleep(helmInstallWaitPeriod)
	err = multierror.Append(err, m.applyCertSecret())
	err = multierror.Append(err, m.applyClusterIssuer())

	if err.Len() > 0 {
		logrus.Errorf("Can not apply the additional cert manager objects: %s", err.Error())
	}
}

func (m *certManager) Uninstall(purge bool) {
	var err *multierror.Error
	client, e := m.contextHandler.GetDynamicClient()
	if e != nil {
		logrus.Errorf("unable to get dynamic client: %s", e)
		return
	}

	m.manager.Uninstall(releaseName, m.namespace, purge)

	clientSet, e := m.contextHandler.GetClientSet()
	if e != nil {
		logrus.Errorf("unable to get k8s client: %s", e)
		return
	}

	e = clientSet.
		CoreV1().
		Secrets(m.namespace).
		Delete(m.ctx, issuerName, metav1.DeleteOptions{})
	err = multierror.Append(err, e)

	e = client.Resource(groupVersion.WithResource("clusterissuers")).
		Delete(m.ctx, issuerName, metav1.DeleteOptions{})
	err = multierror.Append(err, e)

	if err.Len() > 0 {
		logrus.Errorf("Unable to uninstall the certManager plugin: %s", err)
	} else {
		logrus.Info("CertManager plugin successfully uninstalled.")
	}
}

func (m *certManager) Phase() apis.Phase {
	return apis.CLUSTER_TOOLS_INSTALL
}

func (m *certManager) applyCertSecret() error {
	caRoot, e := sh.RunCmd("mkcert", "-CAROOT")
	if e != nil {
		return fmt.Errorf("unable to get the mkcert CA root: %s", e)
	}
	caRoot = strings.Trim(caRoot, "\r\n \t")

	crt, e := ioutil.ReadFile(path.Join(caRoot, "rootCA.pem"))
	if e != nil {
		return fmt.Errorf("unable to read the mkcert RootCA certificate: %s", e)
	}
	key, e := ioutil.ReadFile(path.Join(caRoot, "rootCA-key.pem"))
	if e != nil {
		return fmt.Errorf("unable to read the mkcert RootCA key: %s", e)
	}

	clientSet, e := m.contextHandler.GetClientSet()
	if e != nil {
		return fmt.Errorf("unable to get k8s client: %s", e)
	}
	secretInterface := clientSet.CoreV1().Secrets(m.namespace)

	secret := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Namespace: m.namespace, Name: issuerName},
		Type:       v1.SecretTypeTLS,
		Data: map[string][]byte{
			v1.TLSCertKey:       crt,
			v1.TLSPrivateKeyKey: key,
		},
	}

	_, e = secretInterface.Get(m.ctx, issuerName, metav1.GetOptions{})
	if errors.IsNotFound(e) {
		_, e = secretInterface.Create(m.ctx, secret, metav1.CreateOptions{})
	} else if e == nil {
		_, e = secretInterface.Update(m.ctx, secret, metav1.UpdateOptions{})
	}

	if e != nil {
		return fmt.Errorf("applying the secret failed: %s", e)
	}
	logrus.Debugf("CertSecret '%s' successfully added", issuerName)
	return nil
}

func (m *certManager) applyClusterIssuer() error {
	client, e := m.contextHandler.GetDynamicClient()
	if e != nil {
		return e
	}

	// no namespace needed. The referenced secret will be installed in the same namespace as the cert-manager helm chart.
	issuer := &unstructured.Unstructured{}
	issuer.SetUnstructuredContent(map[string]interface{}{"spec": map[string]interface{}{"ca": map[string]interface{}{"secretName": issuerName}}})
	issuer.SetGroupVersionKind(groupVersion.WithKind("ClusterIssuer"))
	issuer.SetName(issuerName)

	clusterIssuerInterface := client.Resource(groupVersion.WithResource("clusterissuers"))
	old, e := clusterIssuerInterface.Get(m.ctx, issuerName, metav1.GetOptions{})
	if errors.IsNotFound(e) {
		_, e = clusterIssuerInterface.Create(m.ctx, issuer, metav1.CreateOptions{})
	} else if e == nil {
		resourceVersion := old.GetResourceVersion()
		issuer.SetResourceVersion(resourceVersion)
		_, e = clusterIssuerInterface.Update(m.ctx, issuer, metav1.UpdateOptions{})
	}

	if e != nil {
		return fmt.Errorf("applying the cluster issuer failed: %s", e)
	}
	return nil
}
