#!/bin/bash

set -e

PODMAN=podman
FORCE="false"

DEFAULT_BUNDLE_IMAGE="quay.io/wto/web-terminal-operator-metadata:next"
DEFAULT_INDEX_IMAGE="quay.io/wto/web-terminal-operator-index:next"

SCRATCH_DIR=$(mktemp -d)
echo "Using $SCRATCH_DIR for temporary storage"

CSV_FILENAME="web-terminal.clusterserviceversion.yaml"

error() {
  echo "[ERROR] $1"
  exit 1
}

parse_args() {
  while [[ "$#" -gt 0 ]]; do
    case $1 in
      '--bundle-image') BUNDLE_IMAGE="$2"; shift 1;;
      '--index-image') INDEX_IMAGE="$2"; shift 1;;
      '--container-tool') PODMAN="$2"; shift 1;;
      '--force') FORCE="true";;
      *) echo "[ERROR] Unknown parameter is used: $1."; usage; exit 1;;
    esac
    shift 1
  done
}

restore_changes() {
  if [ -f "${SCRATCH_DIR}/${CSV_FILENAME}" ]; then
    cp "${SCRATCH_DIR}/${CSV_FILENAME}" "manifests/${CSV_FILENAME}"
  fi
}

parse_args "$@"

# Set defaults and warn if pushing to main repos in case of accident
BUNDLE_IMAGE="${BUNDLE_IMAGE:-DEFAULT_BUNDLE_IMAGE}"
INDEX_IMAGE="${INDEX_IMAGE:-DEFAULT_INDEX_IMAGE}"

# Check we're not accidentally pushing to the WTO repos
if [ "$BUNDLE_IMAGE" == "$DEFAULT_BUNDLE_IMAGE" ] && [ "$FORCE" != "true" ]; then
  echo -n "Are you sure you want to push $BUNDLE_IMAGE? [y/N] " && read -r ans && [ "${ans:-N}" = y ] || exit 1
fi
if [ "$INDEX_IMAGE" == "$DEFAULT_INDEX_IMAGE" ] && [ "$FORCE" != "true" ]; then
  echo -n "Are you sure you want to push $INDEX_IMAGE? [y/N] " && read -r ans && [ "${ans:-N}" = y ] || exit 1
fi

# Remove replaces field from WTO CSV since we're building a single-version catalog
cp "manifests/${CSV_FILENAME}" "${SCRATCH_DIR}/${CSV_FILENAME}"
trap restore_changes EXIT
yq -Yi 'del(.spec.replaces)' "manifests/${CSV_FILENAME}"

# Build bundle image
echo "Building bundle image $BUNDLE_IMAGE"
$PODMAN build -f ./build/dockerfiles/Dockerfile -t "$BUNDLE_IMAGE" .
$PODMAN push "$BUNDLE_IMAGE"

# Get digest for bundle image we just built
BUNDLE_DIGEST=$($PODMAN inspect "$BUNDLE_IMAGE" | jq ".[].RepoDigests[0]" -r)
echo "Using bundle $BUNDLE_DIGEST in index"

# Add bundle to index
echo "Building index image $INDEX_IMAGE"
opm index add -c "$PODMAN" --bundles "$BUNDLE" --tag "$INDEX_IMAGE"
$PODMAN push "$INDEX_IMAGE"
