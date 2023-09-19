#!/usr/bin/env bash
set -e

parent_path=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )

KIND="${KIND:=kind}"
KUBECTL="${KUBECTL:=kubectl --context kind-rig}"
HELM="${HELM:=helm --kube-context kind-rig}"

# Create kind cluster
${KIND} get clusters | grep "^rig$" || \
    cat <<EOF | ${KIND} create cluster --name rig --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30000
    hostPort: 30000
    listenAddress: "127.0.0.1"
    protocol: TCP
EOF

# Ensure rig-system namespace exists
${KUBECTL} get namespace rig-system || \
    ${KUBECTL} create namespace rig-system

# Install docker registry in rig-system namespace
${KUBECTL} apply -n rig-system \
    -f "${parent_path}/../registry/registry.yaml"

# Ensure required repositories are available
${HELM} repo list | grep '^jetstack\s*https://charts.jetstack.io\s*$' || \
    ${HELM} repo add jetstack https://charts.jetstack.io
${HELM} repo list | grep '^metrics-server\s*https://kubernetes-sigs.github.io/metrics-server\s*$' || \
    ${HELM} repo add metrics-server https://kubernetes-sigs.github.io/metrics-server
${HELM} repo update

# Install cert-manager
${HELM} upgrade --install cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --create-namespace \
    --version v1.13.0 \
    --set installCRDs=true

# Install metrics-server
${HELM} upgrade --install metrics-server metrics-server/metrics-server \
    --namespace kube-system \
    --set args={--kubelet-insecure-tls}
