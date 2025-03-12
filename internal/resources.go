package internal

import (
	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/network"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ipamIpv4 = `{"type": "host-local","ranges": [[{"subnet": "1.1.1.0/24"}]],"dataDir": "/run/my-orchestrator/container-ipam-state"}`

func defineSriovPolicy(name string, sriovDevice string, node string, numVfs int, resourceName string, deviceType string, options ...func(*sriovv1.SriovNetworkNodePolicy)) *sriovv1.SriovNetworkNodePolicy {
	nodePolicy := &sriovv1.SriovNetworkNodePolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: sriovv1.SchemeGroupVersion.String(),
			Kind:       "SriovNetworkNodePolicy",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "openshift-sriov-network-operator",
		},
		Spec: sriovv1.SriovNetworkNodePolicySpec{
			NodeSelector: map[string]string{
				"kubernetes.io/hostname": node,
			},
			NumVfs:       numVfs,
			ResourceName: resourceName,
			Priority:     99,
			NicSelector: sriovv1.SriovNetworkNicSelector{
				PfNames: []string{sriovDevice},
			},
			DeviceType: deviceType,
		},
	}
	for _, o := range options {
		o(nodePolicy)
	}
	return nodePolicy
}

func defineSriovNetwork(name string, namespace string, resourceName string, ipam string, options ...network.SriovNetworkOptions) *sriovv1.SriovNetwork {
	return &sriovv1.SriovNetwork{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "openshift-sriov-network-operator",
		},
		Spec: sriovv1.SriovNetworkSpec{
			ResourceName:     resourceName,
			IPAM:             ipam,
			NetworkNamespace: namespace,
			// Enable the linkState instead of auto so even if the PF is down we can still use the VF
			// for pod to pod connectivity tests in the same host
			LinkState: "enable",
		}}
}

func defineNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"pod-security.kubernetes.io/audit":               "privileged",
				"pod-security.kubernetes.io/enforce":             "privileged",
				"pod-security.kubernetes.io/warn":                "privileged",
				"security.openshift.io/scc.podSecurityLabelSync": "false",
			},
		}}
}
