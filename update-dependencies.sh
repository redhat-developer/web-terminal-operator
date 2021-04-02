#!/bin/bash

set -e

SCRIPT_DIR=${PROJECT_ROOT:-$(cd "$(dirname "$0")" || exit; pwd)}

DEPENDENCIES_DIR="${SCRIPT_DIR}/dependencies"

CRDS_DIR="${DEPENDENCIES_DIR}/devworkspace-crds"
OPERATOR_DIR="${DEPENDENCIES_DIR}/devworkspace-operator"

CRDS_REPO="https://github.com/devfile/kubernetes-api.git"
OPERATOR_REPO="https://github.com/devfile/devworkspace-operator.git"

DEVWORKSPACE_API_VERSION=${DEVWORKSPACE_API_VERSION:-"v1alpha1"}
# 400b4b7 is the https://github.com/devfile/devworkspace-operator/pull/146
# but it's going to be replaced with master or even eventually with v1alpha1 devworkspace operator branch or tag
DEVWORKSPACE_OPERATOR_VERSION=${DEVWORKSPACE_OPERATOR_VERSION:-"400b4b7"}

COMBINED_DIR="${DEPENDENCIES_DIR}/devworkspace-dependencies"

function log() {
  if [ -t 1 ]; then
    echo "$1" | sed 's|.*|\o033[0;31m&\o033[0m|'
  else
    echo "$1"
  fi
}

function update_dep() {
  local repo=$1
  local checkout_path=$2
  local version=$3
  local save_deploy_files=${4:-"false"}
  mkdir -p "$checkout_path"
  pushd "$checkout_path" > /dev/null
  if [ ! -d .git ]; then
    git init
    git remote add origin -f "$repo"
    git config core.sparsecheckout true
    echo "deploy/crds/*" > .git/info/sparse-checkout
    echo "pkg/apis/*" >> .git/info/sparse-checkout
    if [[ $save_deploy_files == "true" ]]; then
      echo "deploy/*.yaml" >> .git/info/sparse-checkout
      echo "deploy/os/*" >> .git/info/sparse-checkout
    fi
  else
    git remote set-url origin "$repo"
  fi
  git fetch --tags -p origin
  if git show-ref --verify "refs/tags/${version}" --quiet; then
    log 'Version is specified from tag'
    git checkout "tags/${version}"
  elif [ ! -z $(git ls-remote --heads origin ${version}) ]; then
    log 'Version is specified from branch'
    git checkout "$version" && git reset --hard "origin/${version}"
  else
    log 'Version is specified from revision'
    git checkout ${version}
  fi
  popd > /dev/null
}

log "Checking out repo '$CRDS_REPO' to '$CRDS_DIR' with version '$DEVWORKSPACE_API_VERSION'"
update_dep "$CRDS_REPO" "$CRDS_DIR" "$DEVWORKSPACE_API_VERSION"

log "Checking out repo '$OPERATOR_REPO' to '$OPERATOR_DIR' with version '$DEVWORKSPACE_OPERATOR_VERSION'"
update_dep "$OPERATOR_REPO" "$OPERATOR_DIR" "$DEVWORKSPACE_OPERATOR_VERSION" "true"

log "Merging repos to $COMBINED_DIR"
rm -rf "$COMBINED_DIR"
mkdir -p "$COMBINED_DIR"
cp -rn "${OPERATOR_DIR}/"* "${COMBINED_DIR}/"
cp -rn "${CRDS_DIR}/"* "${COMBINED_DIR}/"
# OLM automatically provides users permissions for CRs which are brought by the operator
rm -rf "${COMBINED_DIR}/deploy/edit-workspaces-cluster-role.yaml"
rm -rf "${COMBINED_DIR}/deploy/view-workspaces-cluster-role.yaml"
# Don't care about devworkspacetemplates for now
rm -rf "${COMBINED_DIR}/deploy/crds/workspace.devfile.io_devworkspacetemplates_crd.yaml"
