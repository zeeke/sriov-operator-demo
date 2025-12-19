package scenarios

import (
	"testing"

	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/cluster"
	"github.com/stretchr/testify/assert"
)

func TestMellanoxDemoDefaultValues(t *testing.T) {
	discoverSriovFn = discoverSriovFnMellanox

	objects, err := mellanoxDemo()
	assert.NoError(t, err)

	assertGoldenFile(t, objects)
}

func TestMellanoxDemoCustomValues(t *testing.T) {
	discoverSriovFn = discoverSriovFnMellanox

	defer mockEnv(t, "MELLANOX_NICS_APP_NAMESPACE", "mlxapp-ns")()
	defer mockEnv(t, "MELLANOX_NICS_RDMA_RESOURCE_NAME", "mlxrdmacustom")()
	defer mockEnv(t, "MELLANOX_NICS_RDMA_NUM_VFS", "42")()
	defer mockEnv(t, "MELLANOX_NICS_NETDEVICE_RESOURCE_NAME", "mlxnetdevicecustom")()
	defer mockEnv(t, "MELLANOX_NICS_NETDEVICE_NUM_VFS", "43")()
	defer mockEnv(t, "MELLANOX_NICS_IPAM", `'{"type": "host-local","ranges": [[{ "subnet": "10.1.2.0/24" }], [{ "subnet": "2001:db8:1::0/64" }]],"dataDir": "/run/my-orchestrator/container-ipam-state"}`)()

	objects, err := mellanoxDemo()
	assert.NoError(t, err)

	assertGoldenFile(t, objects)
}

func discoverSriovFnMellanox(clients *testclient.ClientSet, operatorNamespace string) (*cluster.EnabledNodes, error) {
	return &cluster.EnabledNodes{
		Nodes: []string{"node1"},
		States: map[string]sriovv1.SriovNetworkNodeState{
			"node1": {
				Status: sriovv1.SriovNetworkNodeStateStatus{
					Interfaces: []sriovv1.InterfaceExt{
						{Name: "ens1f0", DeviceID: "1015", Vendor: "15b3", Driver: "mlx5_core"},
					},
				},
			},
		},
		IsSecureBootEnabled: map[string]bool{
			"node1": false,
		},
	}, nil
}
