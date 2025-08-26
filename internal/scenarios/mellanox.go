package scenarios

import (
	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

func init() {
	Index["mellanox-nics"] = mellanoxDemo
}

type mellanoxNicsConfig struct {
	AppNamespace string       `env:"APP_NAMESPACE, default=demo-mellanox"`
	Rdma         policyConfig `env:",prefix=RDMA_"`
	Netdevice    policyConfig `env:",prefix=NETDEVICE_"`
}

func mellanoxDemo() ([]runtime.Object, error) {
	var c mellanoxNicsConfig
	err := loadConfigFromEnv(&c, "MELLANOX_NICS_")
	if err != nil {
		return nil, err
	}

	if c.Rdma.ResourceName == "" {
		c.Rdma.ResourceName = "mellanoxrdma"
	}
	if c.Netdevice.ResourceName == "" {
		c.Netdevice.ResourceName = "mellanoxnetdevice"
	}

	clients := testclient.New("")

	sriovInfos, err := discoverSriovFn(clients, "openshift-sriov-network-operator")
	if err != nil {
		return nil, err
	}

	var node string = sriovInfos.Nodes[0]

	nic, err := sriovInfos.FindOneMellanoxSriovDevice(node)
	if err != nil {
		return nil, err
	}

	netdevicePolicy := DefineSriovPolicy("demo-mellanox-netdevice", nic.Name+"#10-20", node, c.Netdevice.NumVfs, c.Netdevice.ResourceName, "netdevice")
	rdmaPolicy := DefineSriovPolicy("demo-mellanox-rdma", nic.Name+"#21-31", node, c.Rdma.NumVfs, c.Rdma.ResourceName, "netdevice", func(snnp *sriovv1.SriovNetworkNodePolicy) {
		snnp.Spec.IsRdma = true
	})

	netdeviceNet := DefineSriovNetwork("demo-mellanox-netdevice", c.AppNamespace, c.Netdevice.ResourceName, ipamIpv4)
	rdmaNet := DefineSriovNetwork("demo-mellanox-rdma", c.AppNamespace, c.Rdma.ResourceName, ipamIpv4)

	workloadNs := DefineNamespace(c.AppNamespace)

	deploymentNetdevice := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-mellanox-netdev", Namespace: c.AppNamespace},
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
		ObjectMeta: metav1.ObjectMeta{Name: "demo-mellanox-rdma", Namespace: c.AppNamespace},
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
