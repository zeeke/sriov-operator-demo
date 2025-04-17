package scenarios

import (
	"context"
	"fmt"
	"strings"

	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/pod"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

func getConfigDaemonPod(cs *testclient.ClientSet, nodeName string) (*corev1.Pod, error) {
	pods := &corev1.PodList{}
	label, err := labels.Parse("app=sriov-network-config-daemon")
	if err != nil {
		return nil, fmt.Errorf("can't parse label: %w", err)
	}
	field, err := fields.ParseSelector(fmt.Sprintf("spec.nodeName=%s", nodeName))
	if err != nil {
		return nil, fmt.Errorf("can't parse field selector: %w", err)
	}

	listOptions := &runtimeclient.ListOptions{Namespace: "openshift-sriov-network-operator", LabelSelector: label, FieldSelector: field}
	err = cs.List(context.Background(), pods, listOptions)
	if err != nil {
		return nil, fmt.Errorf("can't list pods with options [%+v]: %w", listOptions, err)
	}
	if len(pods.Items) != 1 {
		return nil, fmt.Errorf("bad number of config-daemon pods for node [%s]: %w", nodeName, err)
	}

	return &pods.Items[0], nil
}

func runCommandOnConfigDaemon(cs *testclient.ClientSet, nodeName string, command ...string) (string, string, error) {
	configDaemonPod, err := getConfigDaemonPod(cs, nodeName)
	if err != nil {
		return "", "", fmt.Errorf("can't get config-daemon pod for node [%s]: %w", nodeName, err)
	}
	output, errOutput, err := pod.ExecCommand(cs, configDaemonPod, command...)
	return output, errOutput, err
}

func findUnusedSriovDevices(cs *testclient.ClientSet, node string, sriovDevices []*sriovv1.InterfaceExt) ([]*sriovv1.InterfaceExt, error) {

	filteredDevices := []*sriovv1.InterfaceExt{}
	stdout, stderr, err := runCommandOnConfigDaemon(cs, node, "ip", "route")
	if err != nil {
		return nil, fmt.Errorf("can't get IP routes for node [%s]: %w\nout:[%s]\nerr[%s]", node, err, stdout, stderr)
	}

	routes := strings.Split(stdout, "\n")

	for _, device := range sriovDevices {
		if isDefaultRouteInterface(device.Name, routes) {
			continue
		}
		stdout, stderr, err := runCommandOnConfigDaemon(cs, node, "ip", "link", "show", device.Name)
		if err != nil {
			fmt.Printf("Can't query link state for device [%s]: %s", device.Name, err.Error())
			continue
		}

		if len(stdout) == 0 {
			fmt.Printf("Can't query link state for device [%s]: stderr:[%s]", device.Name, stderr)
			continue
		}

		if strings.Contains(stdout, "master ovs-system") {
			continue // The interface is not active
		}

		filteredDevices = append(filteredDevices, device)
	}
	if len(filteredDevices) == 0 {
		return nil, fmt.Errorf("unused sriov devices not found")
	}
	return filteredDevices, nil
}

func isDefaultRouteInterface(intfName string, routes []string) bool {
	for _, route := range routes {
		if strings.HasPrefix(route, "default") && strings.Contains(route, "dev "+intfName) {
			return true
		}
	}
	return false
}
