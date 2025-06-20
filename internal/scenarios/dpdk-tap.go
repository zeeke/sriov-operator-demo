package scenarios

import (
	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/cluster"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"

	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
)

func init() {
	Index["dpdk-tap"] = dpdkTapDemo
}

func dpdkTapDemo() ([]runtime.Object, error) {
	clients := testclient.New("")

	sriovInfos, err := cluster.DiscoverSriov(clients, "openshift-sriov-network-operator")
	if err != nil {
		return nil, err
	}

	var node string = sriovInfos.Nodes[0]

	nic, err := FindOneIntelSriovDevice(sriovInfos, node)
	if err != nil {
		return nil, err
	}

	// Create SR-IOV policy for DPDK (VFIO-PCI)
	dpdkPolicy := DefineSriovPolicy("demo-dpdk-vfio", nic.Name+"#10-20", node, 32, "dpdkvfio", "vfio-pci")

	// Create SR-IOV network for DPDK
	dpdkNet := DefineSriovNetwork("demo-dpdk-vfio", "demo-dpdk-tap", "dpdkvfio", "")

	// Create NetworkAttachmentDefinition for tap-cni
	tapNetworkAttachmentDef := &nadv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-tap-cni",
			Namespace: "demo-dpdk-tap",
		},
		Spec: nadv1.NetworkAttachmentDefinitionSpec{
			Config: `{
				"cniVersion": "0.3.1",
				"name": "demo-tap-cni",
				"type": "tap"
			}`,
		},
	}

	// Create namespace
	workloadNs := DefineNamespace("demo-dpdk-tap")

	// Create ServiceAccount for DPDK workloads
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dpdk-serviceaccount",
			Namespace: "demo-dpdk-tap",
		},
	}

	// Create RoleBinding to allow the ServiceAccount to use privileged SCC
	roleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dpdk-privileged-scc-binding",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "dpdk-serviceaccount",
				Namespace: "demo-dpdk-tap",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "system:openshift:scc:privileged",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	// Create sender deployment
	senderDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "dpdk-sender", Namespace: "demo-dpdk-tap"},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "dpdk-sender"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "dpdk-sender"},
					Annotations: map[string]string{
						"k8s.v1.cni.cncf.io/networks": "demo-dpdk-vfio",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "dpdk-serviceaccount",
					NodeSelector: map[string]string{
						"kubernetes.io/hostname": node,
					},
					Containers: []corev1.Container{makeDpdkSenderContainer()},
				},
			},
			Replicas: ptr.To[int32](1),
		},
	}

	// Create receiver deployment
	receiverDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "dpdk-receiver", Namespace: "demo-dpdk-tap"},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "dpdk-receiver"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "dpdk-receiver"},
					Annotations: map[string]string{
						"k8s.v1.cni.cncf.io/networks": `[
							{"name": "demo-dpdk-vfio", "mac": "60:00:00:00:00:02"},
							{"name": "demo-tap-cni", "interface": "tap0"}
						]`,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "dpdk-serviceaccount",
					NodeSelector: map[string]string{
						"kubernetes.io/hostname": node,
					},
					Containers: []corev1.Container{makeDpdkReceiverContainer()},
				},
			},
			Replicas: ptr.To[int32](1),
		},
	}

	return []runtime.Object{
		dpdkPolicy,
		dpdkNet,
		workloadNs,
		serviceAccount,
		roleBinding,
		tapNetworkAttachmentDef,
		senderDeployment,
		receiverDeployment,
	}, nil
}

func makeDpdkSenderContainer() corev1.Container {
	return corev1.Container{
		Name:    "dpdk-sender",
		Image:   "quay.io/openshift-kni/dpdk:4.19",
		Command: []string{"/bin/bash", "-c"},
		Args: []string{`
			dpdk-testpmd --lcores='0@(0-127),1@(0-127)' -a ${PCIDEVICE_OPENSHIFT_IO_DPDKVFIO} --no-huge -m 256M -- --forward-mode=txonly --eth-peer=0,60:00:00:00:00:02 --stats-period=10 --auto-start --total-num-mbufs=2048 || sleep inf
		`},
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:  ptr.To(int64(0)),
			Privileged: ptr.To(true),
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"SYS_ADMIN", "IPC_LOCK", "SYS_RESOURCE", "NET_RAW", "SYS_NICE"},
			},
		},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				"memory": resource.MustParse("1Gi"),
				"cpu":    resource.MustParse("2000m"),
			},
			Requests: corev1.ResourceList{
				"memory": resource.MustParse("1Gi"),
				"cpu":    resource.MustParse("1000m"),
			},
		},
	}
}

func makeDpdkReceiverContainer() corev1.Container {
	return corev1.Container{
		Name:  "dpdk-receiver",
		Image: "quay.io/openshift-kni/dpdk:4.19",

		Command: []string{"/bin/bash", "-c"},
		Args: []string{`
			
			# Start packet capture on tap interface in background
			#tcpdump -i tap0 -l -n | while read line; do
			#	echo "TAP RECEIVED: $line"
			#done &
			
			# Run testpmd to forward packets from DPDK port to tap interface
			dpdk-testpmd --lcores='0@(0-127),1@(0-127)' -a ${PCIDEVICE_OPENSHIFT_IO_DPDKVFIO} --no-huge -m 256M --vdev=net_tap0,iface=tap0 -- --forward-mode=io --auto-start --stats-period=1 --total-num-mbufs=2048 || sleep inf
		`},
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:  ptr.To(int64(0)),
			Privileged: ptr.To(true),
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"SYS_ADMIN", "IPC_LOCK", "SYS_RESOURCE", "NET_RAW", "SYS_NICE"},
			},
		},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				"memory": resource.MustParse("1Gi"),
				"cpu":    resource.MustParse("2"),
			},
			Requests: corev1.ResourceList{
				"memory": resource.MustParse("1Gi"),
				"cpu":    resource.MustParse("2"),
			},
		},
	}
}
