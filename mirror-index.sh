#!/bin/bash

set -e

INDEX_IMG=""
INDEX_MIRROR=""
MIRROR_MANIFEST_DIR="./mirror-index-manifests/"
function parse_arguments() {
    if [[ $# -lt 1 ]]; then print_usage; exit; fi

    while [[ $# -gt 0 ]]; do
        key="$1"
        case $key in
            -i|--index)
            INDEX_IMG=$2
            shift 2
            ;;
            -t|--target)
            INDEX_MIRROR=$2
            shift 2
            ;;
            -v|--verbose)
            set -x
            shift 1
            ;;
            '--help')
            print_usage
            exit 0
            ;;
            *)
            echo -e "Unknown option $1 is specified. See usage:\n"
            print_usage
            exit 0
        esac
    done
}

print_usage () {
  echo "Usage:   $0 -i [ORIGINAL_INDEX] -t [TARGET_INDEX]"
  echo "Options:
  --index,  -i     the index which should be tested
                   To find index see Index Image Location in cvp-test-report of
                   web-terminal-operator-metadata-container-${WANTED_VERSION}
                   http://external-ci-coldstorage.datahub.redhat.com/cvp/cvp-redhat-operator-bundle-image-validation-test/

  --target, -t     the target where mirrored index should be pushed

  --verbose, -v    print the verbose output during execution

  --help,   -h     help
  "
}

parse_arguments "$@"

echo "Preparing mirrored index $INDEX_MIRROR for $INDEX_IMG"

rm -rf "${MIRROR_MANIFEST_DIR:?}/"

# Go through $INDEX_IMG and grab all the mappings
docker pull "$INDEX_IMG"

oc adm catalog mirror --insecure=true --manifests-only --to-manifests=$MIRROR_MANIFEST_DIR "$INDEX_IMG" quay.io/wto

cd $MIRROR_MANIFEST_DIR

# Create the new mapped-images.txt that contains only the web-terminal mappings
grep web-terminal mapping.txt > mapped-images.txt

# Add the actual index into the mapping
DIGEST=$(skopeo inspect "docker://${INDEX_IMG}" | jq -r '.Digest')
INDEX_DIGEST="${INDEX_IMG%:*}@${DIGEST}"

echo "$INDEX_DIGEST"="$INDEX_MIRROR" >> mapped-images.txt

docker tag "$INDEX_IMG" "$INDEX_MIRROR"
docker push "$INDEX_MIRROR"

# Fix the right side of mapped-images.txt. Replace images with their proper upstream counterparts
sed -i -e 's/web-terminal-tech-preview-web-terminal-tooling-rhel8/web-terminal-tooling/g' \
       -e 's/web-terminal-tech-preview-web-terminal-exec-rhel8/web-terminal-exec/g' \
       -e 's/web-terminal-tech-preview-web-terminal-rhel8-operator/web-terminal-operator/g' \
       -e 's/rh-osbs-web-terminal-operator-metadata/web-terminal-operator-metadata/g' \
       mapped-images.txt

# Fix the left side of mapped-images.txt. registry.redhat.io are not available until after the release happens so we defer to redhat proxy
sed -i -e 's|registry.redhat.io/web-terminal-tech-preview/web-terminal-rhel8-operator|registry-proxy.engineering.redhat.com/rh-osbs/web-terminal-operator|g' \
       -e 's|registry.redhat.io/web-terminal-tech-preview/web-terminal-exec-rhel8|registry-proxy.engineering.redhat.com/rh-osbs/web-terminal-exec|g' \
       -e 's|registry.redhat.io/web-terminal-tech-preview/web-terminal-tooling-rhel8|registry-proxy.engineering.redhat.com/rh-osbs/web-terminal-tooling|g' \
       mapped-images.txt

rm -f mapping.txt

yq -yi '.metadata.name = "web-terminal-index-mirror"' imageContentSourcePolicy.yaml

# Remove all unneed content source policies
yq -yi '. | del(.spec.repositoryDigestMirrors[] | select(.source | contains("web-terminal") | not ))' imageContentSourcePolicy.yaml

# Fix the imageContentSourcePolicy to point to the correct images on quay
sed -i 's/web-terminal-tech-preview-web-terminal-tooling-rhel8/web-terminal-tooling/g' imageContentSourcePolicy.yaml
sed -i 's/web-terminal-tech-preview-web-terminal-exec-rhel8/web-terminal-exec/g' imageContentSourcePolicy.yaml
sed -i 's/web-terminal-tech-preview-web-terminal-rhel8-operator/web-terminal-operator/g' imageContentSourcePolicy.yaml
sed -i 's/rh-osbs-web-terminal-operator-metadata/web-terminal-operator-metadata/g' imageContentSourcePolicy.yaml

oc image mirror --insecure=true --filter-by-os=".*" -f mapped-images.txt

echo "Mirrored index for $INDEX_IMG is prepared and pushed to $INDEX_MIRROR"
echo "Note: you should apply imagecontentsourcepolicy.yaml on the cluster before using this index"
