package scenarios

import (
	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/cluster"
	"github.com/openshift-kni/eco-goinfra/pkg/deployment"
	"github.com/openshift-kni/eco-goinfra/pkg/pod"
	"github.com/zeeke/sriov-operator-demo/internal/ecogoinfra"
	multus "gopkg.in/k8snetworkplumbingwg/multus-cni.v4/pkg/types"

	corev1 "k8s.io/api/core/v1"
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

	sleepContainer, err := pod.NewContainerBuilder("sleep", "quay.io/openshift-kni/cnf-tests:4.19", []string{"/bin/bash", "-c", "sleep INF"}).
		WithSecurityContext(&corev1.SecurityContext{}).
		GetContainerCfg()
	if err != nil {
		return nil, err
	}

	deploymentNetdevice := deployment.NewBuilder(
		ecogoinfra.Stub,
		"demo-intel-netdev",
		"demo-intel",
		map[string]string{"app": "demo-intel-netdev"},
		*sleepContainer,
	).
		WithSecondaryNetwork([]*multus.NetworkSelectionElement{pod.StaticAnnotation("demo-intel-netdevice")}).
		WithReplicas(4).
		Definition

	deploymentVfio := deployment.NewBuilder(
		ecogoinfra.Stub,
		"demo-intel-vfio",
		"demo-intel",
		map[string]string{"app": "demo-intel-vfio"},
		*sleepContainer,
	).
		WithSecondaryNetwork([]*multus.NetworkSelectionElement{pod.StaticAnnotation("demo-intel-vfio")}).
		WithReplicas(4).
		Definition

	return []runtime.Object{
		netdevicePolicy, vfioPolicy,
		netdeviceNet, vfioNet,
		workloadNs,
		deploymentNetdevice, deploymentVfio,
	}, nil
}
