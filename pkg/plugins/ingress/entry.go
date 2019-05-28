package ingress

import (
	"k8s.io/api/extensions/v1beta1"
	"reflect"
)

// ingressEntry is a helper structure for intern handling of updated ingresses.
type ingressEntry struct {
	name        string
	namespace   string
	hostNames   []string
	targetIps   []string
	targetHosts []string
}

// String get the ingress name including the namespace
func (e ingressEntry) String() string {
	return e.namespace + "/" + e.name
}

// convertToIngressEntry converts a k8s ingress into the ingress entry by flatten everything.
func convertToIngressEntry(ingress v1beta1.Ingress) ingressEntry {
	return ingressEntry{
		name:        ingress.Name,
		namespace:   ingress.Namespace,
		hostNames:   getHostNames(ingress),
		targetIps:   getLoadBalancerIps(ingress),
		targetHosts: getLoadBalancerHostNames(ingress),
	}
}

// hasTargets check if this ingress entry has at least one target address.
func (e ingressEntry) hasTargets() bool {
	return len(e.targetIps)+len(e.targetHosts) > 0
}

// hasSameTargets check if this ingress entry has the same targets than o.
func (e ingressEntry) hasSameTargets(o ingressEntry) bool {
	return reflect.DeepEqual(e.targetIps, o.targetIps) &&
		reflect.DeepEqual(e.targetHosts, o.targetHosts)
}

// getAddedHostNames compares this with the given o and returns a list of all added host names in this.
func (e ingressEntry) getAddedHostNames(o ingressEntry) []string {
	return difference(e.hostNames, o.hostNames)
}

// getUpdatedHostNames compares this with the given o and returns a list of all potentially updated host names.
func (e ingressEntry) getUpdatedHostNames(o ingressEntry) []string {
	return intersection(e.hostNames, o.hostNames)
}

// getRemovedHostNames  compares this with the given o and returns a list of all removed host names in this.
func (e ingressEntry) getRemovedHostNames(o ingressEntry) []string {
	return difference(o.hostNames, e.hostNames)
}

// getHostNames is a helper function to extract all host names from the given k8s ingress.
func getHostNames(ingress v1beta1.Ingress) []string {
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
func getLoadBalancerIps(ingress v1beta1.Ingress) []string {
	var result []string
	for _, i := range ingress.Status.LoadBalancer.Ingress {
		if i.IP != "" {
			result = append(result, i.IP)
		}
	}
	return result
}

// getLoadBalancerHostNames is a helper function to extract all target host names from the given k8s ingress.
func getLoadBalancerHostNames(ingress v1beta1.Ingress) []string {
	var result []string
	for _, i := range ingress.Status.LoadBalancer.Ingress {
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
