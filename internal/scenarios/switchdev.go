package scenarios

import (
	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/cluster"

	"k8s.io/apimachinery/pkg/runtime"
)

func init() {
	Index["switchdev"] = switchdevDemo
}

func switchdevDemo() ([]runtime.Object, error) {
	clients := testclient.New("")

	sriovInfos, err := cluster.DiscoverSriov(clients, "openshift-sriov-network-operator")
	if err != nil {
		return nil, err
	}

	testNode, interfaces, err := sriovInfos.FindSriovDevicesAndNode()
	if err != nil {
		return nil, err
	}

	// Avoid testing against primary NIC
	interfaces, err = findUnusedSriovDevices(clients, testNode, interfaces)
	if err != nil {
		return nil, err
	}

	ret := []runtime.Object{}

	for _, intf := range interfaces {
		if !doesInterfaceSupportSwitchdev(intf) {
			continue
		}

		resourceName := "swtichdev" + intf.Name
		policy := DefineSriovPolicy("test-switchdev-policy-"+intf.Name, intf.Name, testNode, 8, resourceName, "netdevice", func(snnp *sriovv1.SriovNetworkNodePolicy) {
			snnp.Spec.EswitchMode = "switchdev"
		})
		ret = append(ret, policy)
	}

	return ret, nil
}

func doesInterfaceSupportSwitchdev(intf *sriovv1.InterfaceExt) bool {
	if intf.Driver == "mlx5_core" {
		return true
	}

	if intf.Driver == "ice" {
		return true
	}

	return false
}
