package scenarios

import (
	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"k8s.io/apimachinery/pkg/runtime"
)

func init() {
	Index["intel-nics"] = intelDemo
}

type intelNicsConfig struct {
	AppNamespace string       `env:"APP_NAMESPACE, default=demo-intel"`
	Vfio         policyConfig `env:",prefix=VFIO_"`
	Netdevice    policyConfig `env:",prefix=NETDEVICE_"`
}

func intelDemo() ([]runtime.Object, error) {
	var c intelNicsConfig
	err := loadConfigFromEnv(&c, "INTEL_NICS_")
	if err != nil {
		return nil, err
	}

	if c.Vfio.ResourceName == "" {
		c.Vfio.ResourceName = "intelvfio"
	}
	if c.Netdevice.ResourceName == "" {
		c.Netdevice.ResourceName = "intelnetdevice"
	}

	clients := testclient.New("")

	sriovInfos, err := discoverSriovFn(clients, "openshift-sriov-network-operator")
	if err != nil {
		return nil, err
	}

	var node string = sriovInfos.Nodes[0]

	nic, err := FindOneIntelSriovDevice(sriovInfos, node)
	if err != nil {
		return nil, err
	}

	netdevicePolicy := DefineSriovPolicy("demo-intel-netdevice", nic.Name+"#10-20", node, c.Netdevice.NumVfs, c.Netdevice.ResourceName, "netdevice")
	vfioPolicy := DefineSriovPolicy("demo-intel-vfio", nic.Name+"#21-31", node, c.Vfio.NumVfs, c.Vfio.ResourceName, "vfio-pci")

	netdeviceNet := DefineSriovNetwork("demo-intel-netdevice", c.AppNamespace, c.Netdevice.ResourceName, ipamIpv4)
	vfioNet := DefineSriovNetwork("demo-intel-vfio", c.AppNamespace, c.Vfio.ResourceName, ipamIpv4)

	workloadNs := DefineNamespace(c.AppNamespace)

	deploymentNetdevice := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-intel-netdev", Namespace: c.AppNamespace},
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
		ObjectMeta: metav1.ObjectMeta{Name: "demo-intel-vfio", Namespace: c.AppNamespace},
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
