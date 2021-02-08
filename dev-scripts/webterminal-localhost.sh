#!/bin/bash
#
# Simple bash script to ease testing OpenShift Terminal changes by running bridge locally.
#

set -e

## One time setup:
# sudo mkdir -p /var/run/secrets/kubernetes.io/serviceaccount/
# sudo ln -s /tmp/kube-in-cluster-emulation /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
# sudo ln -s /tmp/kube-in-cluster-emulation /var/run/secrets/kubernetes.io/serviceaccount/token

mkdir -p /tmp/kube-in-cluster-emulation
# Set up
function setup() {
  SA_SECRET=$(oc get sa console -n openshift-console -o yaml | yq -r '.secrets[].name' | grep "console-token-.*")
  oc get secret ${SA_SECRET} -n openshift-console -o json | jq -r '.data["ca.crt"]' | base64 -d > /tmp/kube-in-cluster-emulation/ca.crt
  oc get secret ${SA_SECRET} -n openshift-console -o json | jq -r '.data["token"]'  | base64 -d > /tmp/kube-in-cluster-emulation/token
}

function update_backend() {
  echo "Patching backend"
  sed -i.bak 's/if terminalHost.Scheme != "https"/if false/' ./pkg/terminal/proxy.go
  rm ./pkg/terminal/proxy.go.bak
  git --no-pager diff ./pkg/terminal/proxy.go
  echo "Compiling patched backend"
  ./build-backend.sh
  echo "Reverting patch"
  sed -i.bak 's/if false/if terminalHost.Scheme != "https"/' ./pkg/terminal/proxy.go
  rm ./pkg/terminal/proxy.go.bak
}

function update_frontend() {
  echo "Patching frontend"
  sed -i.bak -e "s|routingClass: 'web-terminal'|routingClass: 'basic'|" \
             -e "s|id: 'redhat-developer/web-terminal/4.5.0'|id: 'redhat-developer/web-terminal-dev/4.5.0'|" \
      ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts
  rm ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts.bak
  git --no-pager diff ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts
  echo "Compiling patched frontend"
  ./build-frontend.sh
  echo "Reverting patch"
  sed -i.bak -e "s|routingClass: 'basic'|routingClass: 'web-terminal'|" \
             -e "s|id: 'redhat-developer/web-terminal-dev/4.5.0'|id: 'redhat-developer/web-terminal/4.5.0'|" \
      ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts
  rm ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts.bak
}

setup
if [[ $1 == "--update-backend" ]]; then
  update_backend
fi
update_frontend

export KUBERNETES_SERVICE_PORT=6443                       
API=$(oc whoami --show-server)
API=${API##*://}
API=${API%%:*}
export KUBERNETES_SERVICE_HOST=${API}

./examples/run-bridge.sh