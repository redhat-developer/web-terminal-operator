#!/bin/bash

set -e

SCRIPT_DIR=$(cd "$(dirname "$0")"; pwd)
REPO_DIR="${SCRIPT_DIR}/.."
OUTPUT_DIR="${REPO_DIR}/generated/internal"

COMBINED_YAML_PATH="${OUTPUT_DIR}/combined.yaml"
OBJECTS_YAML_DIR="${OUTPUT_DIR}/objects"

WEB_TERMINAL_SA_NAME="web-terminal-controller"
DEPLOYMENT_YAML_PATH="$OBJECTS_YAML_DIR/web-terminal-controller.Deployment.yaml"
CLUSTERROLE_YAML_PATH="$OBJECTS_YAML_DIR/web-terminal-controller.ClusterRole.yaml"
CLUSTERROLEBINDING_YAML_PATH="$OBJECTS_YAML_DIR/web-terminal-controller.ClusterRoleBinding.yaml"

rm -rf "$COMBINED_YAML_PATH" "$OBJECTS_YAML_DIR"
mkdir -p "$OBJECTS_YAML_DIR"

# Parse the deployment yaml out of a CSV
# Params:
#   $1 : path to CSV yaml file
parse_deployment() {
  local file="$1"
  yq -y '.spec.install.spec.deployments[] |
  {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
      "name": .name,
      "namespace": "${NAMESPACE}",
      "labels": {
        "app.kubernetes.io/name": "web-terminal-controller",
        "app.kubernetes.io/part-of": "web-terminal-operator"
      }
    }
  } + {"spec": .spec} |
  .spec.template.spec.containers[0].image = "${WTO_IMG}"' "$file" > "$DEPLOYMENT_YAML_PATH"
}

# Parse the ClusterRole out of a CSVOn branch local-testing

parse_clusterrole() {
  local file="$1"
  yq -y '.spec.install.spec.permissions[] |
  {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRole",
    "metadata": {
      "name": "web-terminal-controller-clusterrole",
      "labels": {
        "app.kubernetes.io/name": "web-terminal-controller",
        "app.kubernetes.io/part-of": "web-terminal-operator"
      }
    }
  } + {"rules": .rules}' "$file" > "$CLUSTERROLE_YAML_PATH"
}

# Create ClusterRoleBinding yaml
parse_clusterrolebinding() {
  cat <<EOF > "$CLUSTERROLEBINDING_YAML_PATH"
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: web-terminal-controller-rolebinding
  labels:
    app.kubernetes.io/name: web-terminal-controller
    app.kubernetes.io/part-of: web-terminal-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: web-terminal-controller-clusterrole
subjects:
- kind: ServiceAccount
  name: ${WEB_TERMINAL_SA_NAME}
  namespace: ${NAMESPACE}
EOF
}

combine_yamls() {
  yq -y '.' "$OBJECTS_YAML_DIR"/* > "$COMBINED_YAML_PATH"
}

for file in manifests/*; do
  echo "Processing file $file"
  KIND=$(yq -r '.kind' "$file")
  if [ "$KIND" == "ClusterServiceVersion" ]; then
    parse_deployment "$file"
    parse_clusterrole "$file"
    parse_clusterrolebinding
  else
    NAME=$(yq -r '.metadata.name' "$file")
    yq -y '. * {"metadata": {"namespace": "${NAMESPACE}"}}' "$file" \
      > "${OBJECTS_YAML_DIR}/${NAME}.${KIND}.yaml"
  fi
done

combine_yamls
