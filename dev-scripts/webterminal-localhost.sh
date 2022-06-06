#!/bin/bash
#
# Bash script that helps to get Web Terminal working on the localhost:9000
#

set -e

function parseArgs() {
  if [ $# -eq 0 ]; then
    help
    exit 0
  fi

  while [[ "$#" -gt 0 ]]; do
    case $1 in
        --emulate-in-cluster) DO_EMULATE_IN_CLUSTER="true";;
        --frontend-patch) DO_FRONTEND_PATCH="true";;
        --frontend) DO_FRONTEND="true";;
        --backend) DO_BACKEND="true";;
        --run) DO_EMULATE_IN_CLUSTER="true"; DO_RUN="true";;
        --all) DO_EMULATE_IN_CLUSTER="true"; DO_FRONTEND="true"; DO_BACKEND="true"; DO_RUN="true";;
        --install) DO_INSTALL="true";;
        --uninstall) DO_UNINSTALL="true";;
        --setup-oauth) DO_SETUP_OAUTH="true";;
        -h|--help) help; exit 0 ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
  done
}

function help() {
  echo "
    This scripts helps to get web terminal working with your local changes available on localhost
    Available args:
      --emulate-in-cluster: emulates in-cluster configuration which is needed for backend
      --backend: compiles patched backend to unblock localhost
      --frontend: compiles patched frontend to unblock localhost
      --run: runs bridge (includes setup)
      --all: includes all from above

      --frontend-patch: patches frontend to unblock you to compile and run manually

      --install: propagating this script into PATH.
                 The symlink is propagated into
                 \$(systemd-path user-binaries) -> $(systemd-path user-binaries)
      --uninstall: removing the installed symlink.

      --setup-oauth: prepare the cluster and openshift-console folder to run bridge with OAuth
            More see: https://github.com/openshift/console#openshift-with-authentication

      --help: get this info message

      Usage:
      execute \"./webterminal-localhost.sh --install\" to make it available from any place.

      Then there are mainly two stories:

      -- You are developing frontend --

      1. Compile the patched backend:
          ./webterminal-localhost.sh --backend
      2. Compile the patched frontend and run bridge
          ./webterminal-localhost.sh --frontend --run

      2.* If you want to compile frontend by your own
          ./webterminal-localhost.sh --frontend-patch (do not forget to exlude it before committing)
        and then
          ./webterminal-localhost.sh --run

      -- You are developing backend --

      1. Compile the patched frontend:
          ./webterminal-localhost.sh --frontend
      2. Compile the patched backend and run bridge
          ./webterminal-localhost.sh --backend --run
          or
          ./webterminal-localhost.sh --emulate-in-cluster
          ./examples/run-bridge.sh
"
}

function install() {
  rm -f "$(systemd-path user-binaries)/webterminal-localhost.sh"
  ln -s "$(dirname "$(readlink -f "$0")")/webterminal-localhost.sh" "$(systemd-path user-binaries)/webterminal-localhost.sh"
  echo "Symlink $(systemd-path user-binaries)/webterminal-localhost.sh is created"
  echo "Open a new terminal to get it propagated into path"
}

function uninstall() {
  rm -f "$(systemd-path user-binaries)/webterminal-localhost.sh"
}

function emulate_in_cluster() {
  mkdir -p /tmp/kube-in-cluster-emulation

  # How ca.crt and token are propagated to pods changed in Kubernetes 1.22 and values no longer stored
  # in a secret. Instead we just read the token and ca.crt from a console pod.
  POD_NAME=$(oc get po -l 'app=console,component=ui' -o json | jq -r '.items[0].metadata.name')
  oc exec "${POD_NAME}" -- cat /var/run/secrets/kubernetes.io/serviceaccount/ca.crt > /tmp/kube-in-cluster-emulation/ca.crt
  oc exec "${POD_NAME}" -- cat /var/run/secrets/kubernetes.io/serviceaccount/token > /tmp/kube-in-cluster-emulation/token
  oc exec "${POD_NAME}" -- cat /var/run/secrets/kubernetes.io/serviceaccount/namespace > /tmp/kube-in-cluster-emulation/namespace
  oc exec "${POD_NAME}" -- cat /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt > /tmp/kube-in-cluster-emulation/service-ca.crt

  export KUBERNETES_SERVICE_PORT=6443
  API=$(oc whoami --show-server)
  API=${API##*://}
  API=${API%%:*}
  export KUBERNETES_SERVICE_HOST=${API}

  if [[ ! -L "/var/run/secrets/kubernetes.io/serviceaccount"
    || $( readlink -f /var/run/secrets/kubernetes.io/serviceaccount) != "/tmp/kube-in-cluster-emulation" ]]; then

    echo "Require sudo access to create /var/run/secrets/kubernetes.io/serviceaccount/ directory
      and create symlink to /tmp/kube-in-cluster-emulation"

    sudo rm -f /var/run/secrets/kubernetes.io/serviceaccount && \
      sudo mkdir -p /var/run/secrets/kubernetes.io/ && \
      sudo ln -s /tmp/kube-in-cluster-emulation /var/run/secrets/kubernetes.io/serviceaccount
  fi
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
  patch_frontend
  echo "Compiling patched frontend"
  ./build-frontend.sh
  echo "Reverting patch"
  sed -i.bak -e "s|routingClass: 'basic'|routingClass: 'web-terminal'|" \
             -e "/components:.*web-terminal-exec/d" \
      ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts
  rm ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts.bak
}

function patch_frontend() {
  echo "Patching frontend"
  # shellcheck disable=SC2016
  sed -i.bak -e "s|routingClass: 'web-terminal'|routingClass: 'basic'|" \
             -e '/       name: .web-terminal-exec./{n;n;
                  a\      components: [{name: "web-terminal-exec",container: {command: ["/go/bin/che-machine-exec","--authenticated-user-id","$(DEVWORKSPACE_CREATOR)","--idle-timeout","$(WEB_TERMINAL_IDLE_TIMEOUT)","--pod-selector","controller.devfile.io/devworkspace_id=$(DEVWORKSPACE_ID)","--use-bearer-token",]}}],
                }' \
      ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts
  rm ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts.bak
  git --no-pager diff ./frontend/packages/console-app/src/components/cloud-shell/cloud-shell-utils.ts
}

function set_up_oauth() {
  echo "Creating OAuth client"
  oc process -f examples/console-oauth-client.yaml | oc apply -f -
  oc get oauthclient console-oauth-client -o jsonpath='{.secret}' > examples/console-client-secret

  echo "Fetching CA"
  oc get secrets -n default --field-selector type=kubernetes.io/service-account-token -o json | \
  jq '.items[0].data."ca.crt"' -r | python -m base64 -d > examples/ca.crt
}

parseArgs "$@"

[ -n "$DO_INSTALL" ] && install

[ -n "$DO_UNINSTALL" ] && uninstall

[ -n "$DO_EMULATE_IN_CLUSTER" ] && emulate_in_cluster

[ -n "$DO_BACKEND" ] && update_backend

[ -n "$DO_FRONTEND_PATCH" ] && patch_frontend

[ -n "$DO_FRONTEND" ] && update_frontend

[ -n "$DO_SETUP_OAUTH" ] && set_up_oauth

if [[ -n "$DO_RUN" ]]; then
  echo "Launching ./examples/run-bridge.sh"
  ./examples/run-bridge.sh
fi
