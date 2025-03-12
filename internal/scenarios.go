package internal

import (
	"fmt"
	"os"

	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/cluster"
	"github.com/openshift-kni/eco-goinfra/pkg/deployment"
	"github.com/openshift-kni/eco-goinfra/pkg/pod"
	"github.com/zeeke/sriov-operator-demo/internal/ecogoinfra"
	multus "gopkg.in/k8snetworkplumbingwg/multus-cni.v4/pkg/types"

	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

type scenarioFactory func() ([]runtime.Object, error)

var Scenarios map[string]scenarioFactory = map[string]scenarioFactory{
	"intel-demo": intelDemo,
}

func DumpScenario(factory scenarioFactory) error {

	resources, err := factory()
	if err != nil {
		return err
	}

	scheme := runtime.NewScheme()
	err = sriovv1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	err = corev1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	err = appsv1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	yamlPrinter := printers.NewTypeSetter(scheme).
		ToPrinter(&printers.YAMLPrinter{})

	for _, x := range resources {
		err := yamlPrinter.PrintObj(x, os.Stdout)
		if err != nil {
			return fmt.Errorf("can't serialize object [%#v]: %w", x, err)
		}
	}

	return nil
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

	netdevicePolicy := defineSriovPolicy("demo-intel-netdevice", nic.Name+"#10-20", node, 32, "intelnetdevice", "netdevice")
	vfioPolicy := defineSriovPolicy("demo-intel-vfio", nic.Name+"#21-31", node, 32, "intelvfio", "vfio-pci")

	netdeviceNet := defineSriovNetwork("demo-intel-netdevice", "demo-intel", "intelnetdevice", ipamIpv4)
	vfioNet := defineSriovNetwork("demo-intel-vfio", "demo-intel", "intelvfio", ipamIpv4)

	workloadNs := defineNamespace("demo-intel")

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

	return []runtime.Object{
		netdevicePolicy, vfioPolicy,
		netdeviceNet, vfioNet,
		workloadNs,
		deploymentNetdevice,
	}, nil
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
