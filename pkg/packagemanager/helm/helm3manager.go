package helm

import (
	"context"
	"fmt"
	"os"

	"github.com/kballard/go-shellquote"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/utils"
)

const namespaceArgument = "--namespace"

type helm3Manager struct {
	context kubernetes.ContextHandler
}

func NewHelm3Manager(context kubernetes.ContextHandler) Manager {
	return &helm3Manager{
		context: context,
	}
}

func (h *helm3Manager) Init() error {
	// do nothing, helm3 do not needs a call to 'helm init' as it do not work with tiller.
	return nil
}

func (h *helm3Manager) AddRepository(name string, url string) error {
	_, e := h.runCommand("repo", "add", name, url)

	if e != nil {
		return e
	}
	return nil
}

func (h *helm3Manager) UpdateRepository() error {
	_, e := h.runCommand("repo", "update")

	if e != nil {
		return e
	}
	return nil
}

func (h *helm3Manager) Install(chart string, release string, namespace string, values map[string]interface{}, wait bool) {
	if e := h.ensureNamespaceExists(namespace); e != nil {
		logrus.Errorf("can not ensure that namespace %s exists: %s", namespace, e)
		return
	}

	var args = []string{
		"--install", "--force",
		namespaceArgument, namespace,
		release, chart,
	}
	if wait {
		args = append(args, "--wait")
	}

	flatValues, e := utils.Flatten(values)
	if e != nil {
		logrus.Warnf("Can not flatten values map; Abort to install chart: %s", e)
		return
	}

	for k, v := range flatValues {
		val := shellquote.Join(fmt.Sprintf("%s=%s", k, v))
		args = append(args, "--set", val)
	}

	response, e := h.runCommand("upgrade", args...)
	if e != nil {
		logrus.Errorf("Can not install (%s) helm chart %s into namespace %s:\n%s", e, chart, namespace, response)
		return
	}

	logrus.Infof("Install of helm chart %s as %s/%s was successful.", chart, namespace, release)
	logrus.Debug(response)
}

func (h *helm3Manager) Uninstall(release string, namespace string, purge bool) {
	var e error
	var response string

	if purge {
		response, e = h.runCommand("uninstall", namespaceArgument, namespace, release)
	} else {
		response, e = h.runCommand("uninstall", namespaceArgument, namespace, "--keep-history", release)
	}

	if e != nil {
		logrus.Errorf("Can not delete helm release %s: %s\n%s", release, e, response)
		return
	}
	logrus.Infof("Helm release %s successfully deleted.", release)
	logrus.Debug(response)
}

func (h *helm3Manager) GetVersion() string {
	return "3"
}

func (h *helm3Manager) runCommand(command string, args ...string) (string, error) {
	prefix := []string{command}
	if h.context.GetContextName() != "" {
		prefix = append(prefix, "--kube-context", h.context.GetContextName())
	}
	if h.context.GetConfigFile() != "" {
		prefix = append(prefix, "--kubeconfig", h.context.GetConfigFile())
	}

	cmd := sh.ExecCommand("helm", append(prefix, args...)...)
	cmd.Env = append(cmd.Env, os.Environ()...)

	bytes, e := cmd.CombinedOutput()
	output := string(bytes)
	if e != nil {
		return output, fmt.Errorf("run helm %s was not successful: %s", command, e)
	}
	return output, nil
}

func (h *helm3Manager) ensureNamespaceExists(namespace string) error {
	ctx := context.Background()
	clientSet, e := h.context.GetClientSet()
	if e != nil {
		return e
	}
	logrus.Debugf("Check if namespace '%s' exits.", namespace)
	ns, e := clientSet.CoreV1().
		Namespaces().
		Get(ctx, namespace, metav1.GetOptions{})

	if e == nil {
		logrus.Tracef("Namespace '%s' exits: %s", namespace, ns)
		return nil
	}

	if !errors.IsNotFound(e) {
		return e
	}

	logrus.Debugf("Creating namespace '%s'.", namespace)
	_, e = clientSet.CoreV1().
		Namespaces().
		Create(
			ctx,
			&v1.Namespace{
				TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{Name: namespace},
			},
			metav1.CreateOptions{},
		)
	return e
}
