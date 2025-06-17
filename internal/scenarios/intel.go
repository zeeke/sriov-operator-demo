package scenarios

import (
	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/cluster"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"k8s.io/apimachinery/pkg/runtime"
)

func init() {
	Index["intel-nics"] = intelDemo
}

func intelDemo() ([]runtime.Object, error) {
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

	netdevicePolicy := DefineSriovPolicy("demo-intel-netdevice", nic.Name+"#10-20", node, 32, "intelnetdevice", "netdevice")
	vfioPolicy := DefineSriovPolicy("demo-intel-vfio", nic.Name+"#21-31", node, 32, "intelvfio", "vfio-pci")

	netdeviceNet := DefineSriovNetwork("demo-intel-netdevice", "demo-intel", "intelnetdevice", ipamIpv4)
	vfioNet := DefineSriovNetwork("demo-intel-vfio", "demo-intel", "intelvfio", ipamIpv4)

	workloadNs := DefineNamespace("demo-intel")

	deploymentNetdevice := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-intel-netdev", Namespace: "demo-intel"},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "demo-intel-netdev"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": "demo-intel-netdev"},
					Annotations: map[string]string{"k8s.v1.cni.cncf.io/networks": string("demo-intel-netdev")},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{makeSleepContainer()},
				},
			},
			Replicas: ptr.To[int32](4),
		},
	}

	deploymentVfio := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-intel-vfio", Namespace: "demo-intel"},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "demo-intel-vfio"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": "demo-intel-vfio"},
					Annotations: map[string]string{"k8s.v1.cni.cncf.io/networks": string("demo-intel-vfio")},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{makeSleepContainer()},
				},
			},
			Replicas: ptr.To[int32](4),
		},
	}

	return []runtime.Object{
		netdevicePolicy, vfioPolicy,
		netdeviceNet, vfioNet,
		workloadNs,
		deploymentNetdevice, deploymentVfio,
	}, nil
}
