package scenarios

import (
	"fmt"

	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/cluster"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/network"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const ipamIpv4 = `{"type": "host-local","ranges": [[{"subnet": "1.1.1.0/24"}]],"dataDir": "/run/my-orchestrator/container-ipam-state"}`

var (
	capabilityAll          = []corev1.Capability{"ALL"}
	defaultGroupID         = int64(3000)
	defaultUserID          = int64(2000)
	defaultSecurityContext = &corev1.SecurityContext{
		AllowPrivilegeEscalation: ptr.To(false),
		RunAsNonRoot:             ptr.To(true),
		SeccompProfile:           &corev1.SeccompProfile{Type: "RuntimeDefault"},
		Capabilities: &corev1.Capabilities{
			Drop: capabilityAll,
		},
		RunAsGroup: &defaultGroupID,
		RunAsUser:  &defaultUserID,
	}
)

func DefineSriovPolicy(name string, sriovDevice string, node string, numVfs int, resourceName string, deviceType string, options ...func(*sriovv1.SriovNetworkNodePolicy)) *sriovv1.SriovNetworkNodePolicy {
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

func DefineSriovNetwork(name string, namespace string, resourceName string, ipam string, options ...network.SriovNetworkOptions) *sriovv1.SriovNetwork {
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

func DefineNamespace(name string) *corev1.Namespace {
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

func FindOneIntelSriovDevice(n *cluster.EnabledNodes, node string) (*sriovv1.InterfaceExt, error) {
	s, ok := n.States[node]
	if !ok {
		return nil, fmt.Errorf("node %s not found", node)
	}

	for _, itf := range s.Status.Interfaces {
		if itf.Vendor == "8086" && sriovv1.IsSupportedModel(itf.Vendor, itf.DeviceID) {
			return &itf, nil
		}
	}

	return nil, fmt.Errorf("unable to find an Intel sriov devices in node %s", node)
}

func makeSleepContainer() corev1.Container {
	return corev1.Container{
		Name:            "sleep",
		Image:           "quay.io/openshift-kni/cnf-tests:4.19",
		Command:         []string{"/bin/bash", "-c", "sleep INF"},
		SecurityContext: &corev1.SecurityContext{},
	}
}
