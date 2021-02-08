#!/bin/bash
#
# Simple bash script to ease testing OpenShift Terminal changes by running bridge locally.
#

set -e

## One time setup:
# sudo ln -s ~/kube-in-cluster-emulation/ca.crt /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
# sudo ln -s ~/kube-in-cluster-emulation/token /var/run/secrets/kubernetes.io/serviceaccount/token

# Set up
function setup() {
  SA_SECRET=$(oc get sa console -n openshift-console -o yaml | yq -r '.secrets[].name' | grep "console-token-.*")
  oc get secret ${SA_SECRET} -n openshift-console -o json | jq -r '.data["ca.crt"]' | base64 -d > ~/kube-in-cluster-emulation/ca.crt
  oc get secret ${SA_SECRET} -n openshift-console -o json | jq -r '.data["token"]'  | base64 -d > ~/kube-in-cluster-emulation/token
}

function update_backend() {
  sed -i.bak 's/if terminalHost.Scheme != "https"/if false/' ./pkg/terminal/proxy.go
  rm ./pkg/terminal/proxy.go.bak
  ./build-backend.sh
  sed -i.bak 's/if false/if terminalHost.Scheme != "https"/' ./pkg/terminal/proxy.go
  rm ./pkg/terminal/proxy.go.bak
}

function update_frontend() {
  sed -i.bak -e "s|routingClass: 'openshift-terminal'|routingClass: 'basic'|" \
             -e "s|id: 'che-incubator/command-line-terminal/4.5.0'|id: 'che-incubator/command-line-terminal-dev/4.5.0'|" \
      ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts
  rm ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts.bak
  ./build-frontend.sh
  sed -i.bak -e "s|routingClass: 'basic'|routingClass: 'openshift-terminal'|" \
             -e "s|id: 'che-incubator/command-line-terminal-dev/4.5.0'|id: 'che-incubator/command-line-terminal/4.5.0'|" \
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