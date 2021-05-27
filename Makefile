SHELL := bash
.SHELLFLAGS = -ec

WTO_IMG ?= quay.io/wto/web-terminal-operator:latest
BUNDLE_IMG ?= quay.io/wto/web-terminal-operator-metadata:next
INDEX_IMG ?= quay.io/wto/web-terminal-operator-index:next
PRODUCTION_ENABLED ?= false
LATEST_INDEX_IMG ?= quay.io/wto/web-terminal-operator-index:latest
GET_DIGEST_WITH ?= skopeo

DOCKER ?= docker

.ONESHELL:
ifndef VERBOSE
MAKEFLAGS += --silent
endif

include build/makefiles/deployment.mk
include build/makefiles/version.mk

all: help

_print_vars:
	echo "Current env vars:"
	echo "    WTO_IMG=$(WTO_IMG)"
	echo "    BUNDLE_IMG=$(BUNDLE_IMG)"
	echo "    INDEX_IMG=$(INDEX_IMG)"
	echo "    LATEST_INDEX_IMG=$(LATEST_INDEX_IMG)"

### build: build the terminal bundle and index and push them to a docker registry
build: _print_vars _check_imgs_env _check_skopeo_installed
	# Create the bundle and push it to a docker registry
	$(DOCKER) build -f ./build/dockerfiles/Dockerfile -t $(BUNDLE_IMG) .
	$(DOCKER) push $(BUNDLE_IMG)

	BUNDLE_DIGEST=$$(skopeo inspect docker://$(BUNDLE_IMG) | jq -r '.Digest')
	BUNDLE_IMG=$(BUNDLE_IMG)
	BUNDLE_IMG_DIGEST="$${BUNDLE_IMG%:*}@$${BUNDLE_DIGEST}"
	# create / update and push an index that contains the bundle
	opm index add -c $(DOCKER) --bundles $${BUNDLE_IMG_DIGEST} --tag $(INDEX_IMG) --from-index $(LATEST_INDEX_IMG)
	$(DOCKER) push $(INDEX_IMG)

### export: export the bundles stored in the index to the exported-manifests folder
export: _print_vars _check_imgs_env
	rm -rf ./generated/exported-manifests
	# Export the bundles with the name web-terminal inside of $(INDEX_IMG)
	# This command basic exports the index back into the old format
	opm index export -c $(DOCKER) -f ./generated/exported-manifests -i $(INDEX_IMG) -o web-terminal

### register_catalogsource: creates the catalogsource to make the operator be available on the marketplace. Image referenced by INDEX_IMG must be pushed and publicly available
register_catalogsource: _print_vars _check_imgs_env _check_skopeo_installed

ifeq ($(GET_DIGEST_WITH),skopeo)
	INDEX_DIGEST=$$(skopeo inspect docker://$(INDEX_IMG) | jq -r '.Digest')
	INDEX_IMG=$(INDEX_IMG)
	INDEX_IMG_DIGEST="$${INDEX_IMG%:*}@$${INDEX_DIGEST}"
else ifeq ($(GET_DIGEST_WITH),$(filter $(GET_DIGEST_WITH),podman docker))   
	$(GET_DIGEST_WITH) pull $(INDEX_IMG)
	INDEX_IMG_DIGEST=$$($(GET_DIGEST_WITH) inspect $(INDEX_IMG) | jq ".[].RepoDigests[0]" -r) 
else
	echo "unsupported GET_DIGEST_WITH is configured"
	exit 1
endif

	# replace references of catalogsource img with your image
	sed -i.bak -e "s|quay.io/wto/web-terminal-operator-index:next|$${INDEX_IMG_DIGEST}|g" ./catalog-source.yaml
	echo ">>>>>>>catalogsource content:>>>>>>>"
	cat ./catalog-source.yaml
	echo ">>>>>>>end of content:>>>>>>>"
	# use ';' to make sure we undo changes to catalog-source.yaml even if command fails.
	oc apply -f ./catalog-source.yaml ; \
	  mv ./catalog-source.yaml.bak ./catalog-source.yaml
	oc apply -f ./imageContentSourcePolicy.yaml

### unregister_catalogsource: unregister the catalogsource and delete the imageContentSourcePolicy
unregister_catalogsource:
	oc delete catalogsource custom-web-terminal-catalog -n openshift-marketplace --ignore-not-found
	oc delete imagecontentsourcepolicy web-terminal-brew-registry-mirror --ignore-not-found

### build_install: build the catalog and create catalogsource and operator subscription on the cluster
build_install: _print_vars _select_controller_image build _reset_controller_image install

### install: creates catalog source along with operator subscription on the cluster
install: _print_vars register_catalogsource
	oc apply -f ./operator-subscription.yaml

### uninstall: uninstalls the Web Terminal Operator in the proper way described in documentation
uninstall:
	kubectl delete subscriptions.operators.coreos.com web-terminal -n openshift-operators --ignore-not-found
	export WTO_CSV=$$(kubectl get csv -o=json | jq -r '[.items[] | select (.metadata.name | contains("web-terminal.v1"))][0].metadata.name') ;\
		kubectl delete csv $${WTO_CSV} -n openshift-operators

_check_imgs_env:
ifndef BUNDLE_IMG
	$(error "BUNDLE_IMG not set")
endif
ifndef INDEX_IMG
	$(error "INDEX_IMG not set")
endif

_check_skopeo_installed:
ifeq ($(shell command -v skopeo 2> /dev/null),)
	$(error "skopeo is required for building and deploying bundle, but is not installed")
endif

.PHONY: help
### help: print this message
help: Makefile
	echo 'Available rules:'
	sed -n 's/^### /    /p' $< | awk 'BEGIN { FS=":" } { printf "%-34s -%s\n", $$1, $$2 }'
	echo ''
	echo 'Supported environment variables:'
	echo '    BUNDLE_IMG                     - The name of the olm registry bundle image. Set to $(BUNDLE_IMG)'
	echo '    INDEX_IMG                      - The name of the olm registry index image. Set to $(INDEX_IMG)'
	echo '    LATEST_INDEX_IMG               - The name of the latest released index. Set to $(LATEST_INDEX_IMG)'
	echo '    PRODUCTION_ENABLED             - If you want to use production images. Set to $(PRODUCTION_ENABLED)'
	echo '    DEVWORKSPACE_API_VERSION       - Branch or tag of the github.com/devfile/kubernetes-api to depend on.'
	echo '    DEVWORKSPACE_OPERATOR_VERSION  - The branch/tag of the terminal manifests.'
	echo '    GET_DIGEST_WITH                - The tool name for obtaining an image didgest. Supported tools: skopeo, podman, docker'
