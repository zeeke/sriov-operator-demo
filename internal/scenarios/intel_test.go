package scenarios

import (
	"testing"

	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/cluster"
	"github.com/stretchr/testify/assert"
)

func TestIntelDemoDefaultValues(t *testing.T) {
	discoverSriovFn = discoverSriovFnIntel

	objects, err := intelDemo()
	assert.NoError(t, err)

	assertGoldenFile(t, objects)
}

func TestIntelDemoCustomValues(t *testing.T) {
	discoverSriovFn = discoverSriovFnIntel

	defer mockEnv(t, "INTEL_NICS_APP_NAMESPACE", "intelapp-ns")()
	defer mockEnv(t, "INTEL_NICS_VFIO_RESOURCE_NAME", "intelvfiocustom")()
	defer mockEnv(t, "INTEL_NICS_VFIO_NUM_VFS", "42")()
	defer mockEnv(t, "INTEL_NICS_NETDEVICE_RESOURCE_NAME", "intelnetdevicecustom")()
	defer mockEnv(t, "INTEL_NICS_NETDEVICE_NUM_VFS", "43")()
	defer mockEnv(t, "INTEL_NICS_IPAM", `'{"type": "host-local","ranges": [[{ "subnet": "10.1.2.0/24" }], [{ "subnet": "2001:db8:1::0/64" }]],"dataDir": "/run/my-orchestrator/container-ipam-state"}`)()

	objects, err := intelDemo()
	assert.NoError(t, err)

	assertGoldenFile(t, objects)
}

func discoverSriovFnIntel(clients *testclient.ClientSet, operatorNamespace string) (*cluster.EnabledNodes, error) {
	return &cluster.EnabledNodes{
		Nodes: []string{"node1"},
		States: map[string]sriovv1.SriovNetworkNodeState{
			"node1": {
				Status: sriovv1.SriovNetworkNodeStateStatus{
					Interfaces: []sriovv1.InterfaceExt{
						{Name: "ens1f0", DeviceID: "158a", Vendor: "8086", Driver: "i40e"},
					},
				},
			},
		},
		IsSecureBootEnabled: map[string]bool{
			"node1": false,
		},
	}, nil
}
