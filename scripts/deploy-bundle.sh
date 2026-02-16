#!/bin/bash

bundle_image=$1

uninstall_operator() {
    oc delete -n openshift-sriov-network-operator sriovnetworknodepolicy --all
    
    # Wait for a stable empty SR-IOV configuration
    oc get crd sriovnetworknodestates.sriovnetwork.openshift.io > /dev/null && \
        until oc get sriovnetworknodestates.sriovnetwork.openshift.io -A -o jsonpath='{.items[*].status.syncStatus}' | grep -qx Failed; do echo "waiting cluster stable"; sleep 5; done
    
    oc delete -n openshift-sriov-network-operator sriovnetwork --all
    oc delete -n openshift-sriov-network-operator sriovibnetwork --all
    oc delete -n openshift-sriov-network-operator sriovoperatorconfig default
    sleep 5

    oc delete crd sriovibnetworks.sriovnetwork.openshift.io
    oc delete crd sriovnetworknodepolicies.sriovnetwork.openshift.io
    oc delete crd sriovnetworknodestates.sriovnetwork.openshift.io
    oc delete crd sriovnetworkpoolconfigs.sriovnetwork.openshift.io
    oc delete crd sriovnetworks.sriovnetwork.openshift.io
    oc delete crd sriovoperatorconfigs.sriovnetwork.openshift.io

    sleep 5

    oc delete mutatingwebhookconfigurations network-resources-injector-config
    oc delete MutatingWebhookConfiguration sriov-operator-webhook-config
    oc delete ValidatingWebhookConfiguration sriov-operator-webhook-config

    oc delete namespace openshift-sriov-network-operator
    sleep 5

    oc annotate node --all sriovnetwork.openshift.io/state-
    oc annotate node --all sriovnetwork.openshift.io/desired-state-
    oc annotate node --all sriovnetwork.openshift.io/current-state-

    oc adm uncordon -l node-role.kubernetes.io/worker=
}

create_namespace() {
    cat << EOF | oc create -f -
apiVersion: v1
kind: Namespace
metadata:
  name: openshift-sriov-network-operator
  annotations:
    workload.openshift.io/allowed: management
  labels:
    pod-security.kubernetes.io/audit: privileged
    pod-security.kubernetes.io/enforce: privileged
    pod-security.kubernetes.io/warn: privileged
EOF
}

create_operator_config() {
    cat <<EOF | oc create -f -
apiVersion: sriovnetwork.openshift.io/v1
kind: SriovOperatorConfig
metadata:
  name: default
  namespace: openshift-sriov-network-operator
spec:
  enableInjector: true
  enableOperatorWebhook: true
  logLevel: 2
  disableDrain: false
EOF
}


uninstall_operator
create_namespace

operator-sdk run bundle \
    ${bundle_image} \
    --namespace openshift-sriov-network-operator

create_operator_config
