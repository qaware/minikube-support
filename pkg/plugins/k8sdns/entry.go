package k8sdns

import (
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
)

// entry is a helper structure for intern handling of updated ingresses and services.
type entry struct {
	name        string
	namespace   string
	typ         string
	hostNames   []string
	targetIps   []string
	targetHosts []string
}

// String get the ingress name including the namespace
func (e entry) String() string {
	return e.namespace + "/" + e.name
}

// hasTargets check if this ingress entry has at least one target address.
func (e entry) hasTargets() bool {
	return len(e.targetIps)+len(e.targetHosts) > 0
}

// hasSameTargets check if this ingress entry has the same targets than o.
func (e entry) hasSameTargets(o *entry) bool {
	return reflect.DeepEqual(e.targetIps, o.targetIps) &&
		reflect.DeepEqual(e.targetHosts, o.targetHosts)
}

// getAddedHostNames compares this with the given o and returns a list of all added host names in this.
func (e entry) getAddedHostNames(o *entry) []string {
	return difference(e.hostNames, o.hostNames)
}

// getUpdatedHostNames compares this with the given o and returns a list of all potentially updated host names.
func (e entry) getUpdatedHostNames(o *entry) []string {
	return intersection(e.hostNames, o.hostNames)
}

// getRemovedHostNames  compares this with the given o and returns a list of all removed host names in this.
func (e entry) getRemovedHostNames(o *entry) []string {
	return difference(o.hostNames, e.hostNames)
}

// getHostNames is a helper function to extract all host names from the given k8s ingress.
func getHostNames(ingress *v1beta1.Ingress) []string {
	hostMap := make(map[string]bool)

	for _, rule := range ingress.Spec.Rules {
		hostMap[rule.Host] = true
	}
	for _, tlsRule := range ingress.Spec.TLS {
		for _, host := range tlsRule.Hosts {
			hostMap[host] = true
		}
	}

	var result []string
	for host := range hostMap {
		result = append(result, host)
	}
	return result
}

// getLoadBalancerIps is a helper function to extract all target ip addresses from the given k8s ingress.
func getLoadBalancerIps(status v1.LoadBalancerStatus) []string {
	var result []string
	for _, i := range status.Ingress {
		if i.IP != "" {
			result = append(result, i.IP)
		}
	}
	return result
}

// getLoadBalancerHostNames is a helper function to extract all target host names from the given k8s ingress.
func getLoadBalancerHostNames(status v1.LoadBalancerStatus) []string {
	var result []string
	for _, i := range status.Ingress {
		if i.Hostname != "" {
			result = append(result, i.Hostname)
		}
	}
	return result
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// intersection returns the elements in `a` that are also in `b`.
func intersection(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var intersection []string
	for _, x := range a {
		if _, found := mb[x]; found {
			intersection = append(intersection, x)
		}
	}
	return intersection
}
