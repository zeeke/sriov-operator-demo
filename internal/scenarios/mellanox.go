package scenarios

import (
	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
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

	sleepContainer, err := pod.NewContainerBuilder("sleep", "quay.io/openshift-kni/cnf-tests:4.19", []string{"/bin/bash", "-c", "sleep INF"}).
		WithSecurityContext(&corev1.SecurityContext{}).
		GetContainerCfg()
	if err != nil {
		return nil, err
	}

	deploymentNetdevice := deployment.NewBuilder(
		ecogoinfra.Stub,
		"demo-mellanox-netdev",
		"demo-mellanox",
		map[string]string{"app": "demo-mellanox-netdev"},
		*sleepContainer,
	).
		WithSecondaryNetwork([]*multus.NetworkSelectionElement{pod.StaticAnnotation("demo-mellanox-netdevice")}).
		WithReplicas(4).
		Definition

	deploymentRdma := deployment.NewBuilder(
		ecogoinfra.Stub,
		"demo-mellanox-rmda",
		"demo-mellanox",
		map[string]string{"app": "demo-mellanox-rdma"},
		*sleepContainer,
	).
		WithSecondaryNetwork([]*multus.NetworkSelectionElement{pod.StaticAnnotation("demo-mellanox-rdma")}).
		WithReplicas(4).
		Definition

	return []runtime.Object{
		netdevicePolicy, rdmaPolicy,
		netdeviceNet, rdmaNet,
		workloadNs,
		deploymentNetdevice, deploymentRdma,
	}, nil
}
