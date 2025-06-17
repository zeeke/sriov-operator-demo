package scenarios

import (
	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/cluster"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

func init() {
	Index["mellanox-nics"] = mellanoxDemo
}

func mellanoxDemo() ([]runtime.Object, error) {
	clients := testclient.New("")

	sriovInfos, err := cluster.DiscoverSriov(clients, "openshift-sriov-network-operator")
	if err != nil {
		return nil, err
	}

	var node string = sriovInfos.Nodes[0]

	nic, err := sriovInfos.FindOneMellanoxSriovDevice(node)
	if err != nil {
		return nil, err
	}

	netdevicePolicy := DefineSriovPolicy("demo-mellanox-netdevice", nic.Name+"#10-20", node, 32, "mellanoxnetdevice", "netdevice")
	rdmaPolicy := DefineSriovPolicy("demo-mellanox-rdma", nic.Name+"#21-31", node, 32, "mallanoxrdma", "netdevice", func(snnp *sriovv1.SriovNetworkNodePolicy) {
		snnp.Spec.IsRdma = true
	})

	netdeviceNet := DefineSriovNetwork("demo-mellanox-netdevice", "demo-mellanox", "mellanoxnetdevice", ipamIpv4)
	rdmaNet := DefineSriovNetwork("demo-mellanox-rdma", "demo-mellanox", "mallanoxrdma", ipamIpv4)

	workloadNs := DefineNamespace("demo-mellanox")

	deploymentNetdevice := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-mellanox-netdev", Namespace: "demo-mellanox"},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "demo-mellanox-netdev"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": "demo-mellanox-netdev"},
					Annotations: map[string]string{"k8s.v1.cni.cncf.io/networks": string("demo-mellanox-netdev")},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{makeSleepContainer()},
				},
			},
			Replicas: ptr.To[int32](4),
		},
	}

	deploymentRdma := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-mellanox-rdma", Namespace: "demo-mellanox"},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "demo-mellanox-rdma"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": "demo-mellanox-rdma"},
					Annotations: map[string]string{"k8s.v1.cni.cncf.io/networks": string("demo-mellanox-rdma")},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{makeSleepContainer()},
				},
			},
			Replicas: ptr.To[int32](4),
		},
	}

	return []runtime.Object{
		netdevicePolicy, rdmaPolicy,
		netdeviceNet, rdmaNet,
		workloadNs,
		deploymentNetdevice, deploymentRdma,
	}, nil
}
