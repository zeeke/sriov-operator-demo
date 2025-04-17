#!/bin/bash

script_dir=$(dirname "$(readlink -f "$0")")

cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraMounts: 
    - hostPath: ${script_dir}/hostname_master-0
      containerPath: /etc/hostname
  - role: worker
    extraMounts: 
    - hostPath: ${script_dir}/hostname_worker-0
      containerPath: /etc/hostname
    #- hostPath: ${script_dir}/secure-boot
    #  containerPath: /sys/kernel/security/lockdown
  - role: worker
    extraMounts: 
    - hostPath: ${script_dir}/hostname_worker-1
      containerPath: /etc/hostname
    #- hostPath: ${script_dir}/secure-boot
    #  containerPath: /sys/kernel/security/lockdown
EOF

kubectl kustomize https://github.com/k8snetworkplumbingwg/sriov-network-operator.git/config/crd/ | kubectl apply -f -
kubectl apply -f examples/mocks/namespaces.yaml
kubectl apply -f examples/mocks/configmaps.yaml
kubectl apply -f examples/mocks/sriovnetworknodestates.yaml
kubectl apply -f examples/mocks/sriovnetworknodestates.yaml --subresource status --server-side

