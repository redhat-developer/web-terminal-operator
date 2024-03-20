SHELL := bash
.SHELLFLAGS = -ec
###########
WTO_IMG ?= quay.io/wto/web-terminal-operator:next
BUNDLE_IMG ?= quay.io/wto/web-terminal-operator-metadata:next
INDEX_IMG ?= quay.io/wto/web-terminal-operator-index:next
GET_DIGEST_WITH ?= skopeo

DOCKER ?= docker

.ONESHELL:
ifndef VERBOSE
MAKEFLAGS += --silent
endif

include build/makefiles/controller.mk
include build/makefiles/version.mk

all: help

_print_vars:
	echo "Current env vars:"
	echo "    WTO_IMG=$(WTO_IMG)"
	echo "    BUNDLE_IMG=$(BUNDLE_IMG)"
	echo "    INDEX_IMG=$(INDEX_IMG)"

### build: build the terminal bundle and index and push them to a docker registry
build: _print_vars _check_imgs_env _check_skopeo_installed
	build/scripts/build_index_image.sh \
		--bundle-image $(BUNDLE_IMG) \
		--index-image $(INDEX_IMG) \
		--container-tool $(DOCKER)

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

### uninstall: uninstalls the Web Terminal Operator Subscription and related ClusterServiceVersion
uninstall:
	# 1. Ensure that all DevWorkspace Custom Resources are removed to avoid issues with finalizers
	# make sure depending objects are clean up as well
	kubectl delete devworkspaces.workspace.devfile.io --all-namespaces --all --wait
	kubectl delete workspaceroutings.controller.devfile.io --all-namespaces --all --wait
	kubectl delete components.controller.devfile.io --all-namespaces --all --wait
	# 2. Uninstall the Operator
	kubectl delete subscriptions.operators.coreos.com web-terminal -n openshift-operators --ignore-not-found
	$(eval WTO_CSV := $(shell kubectl get csv -o=json | jq -r '[.items[] | select (.metadata.name | contains("web-terminal.v1"))][0].metadata.name'))
	kubectl delete csv ${WTO_CSV} -n openshift-operators
	# 3. Remove CRDs
	kubectl delete customresourcedefinitions.apiextensions.k8s.io workspaceroutings.controller.devfile.io
	kubectl delete customresourcedefinitions.apiextensions.k8s.io components.controller.devfile.io
	kubectl delete customresourcedefinitions.apiextensions.k8s.io devworkspaces.workspace.devfile.io
	# 4. Remove DevWorkspace Webhook Server Deployment itself
	kubectl delete deployment/devworkspace-webhook-server -n openshift-operators
	# 5. Remove lingering service, secrets, and configmaps
	kubectl delete all --selector app.kubernetes.io/part-of=devworkspace-operator,app.kubernetes.io/name=devworkspace-webhook-server
	kubectl delete serviceaccounts devworkspace-webhook-server -n openshift-operators
	kubectl delete configmap devworkspace-controller -n openshift-operators
	kubectl delete clusterrole devworkspace-webhook-server
	kubectl delete clusterrolebinding devworkspace-webhook-server
	# 6. Remove mutating/validating webhooks configuration
	kubectl delete mutatingwebhookconfigurations controller.devfile.io
	kubectl delete validatingwebhookconfigurations controller.devfile.io

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
	sed -n 's/^### /    /p' $(MAKEFILE_LIST) | awk 'BEGIN { FS=":" } { printf "%-34s -%s\n", $$1, $$2 }'
	echo ''
	echo 'Supported environment variables:'
	echo '    WTO_IMG                        - The name of the controller image. Set to $(WTO_IMG)'
	echo '    BUNDLE_IMG                     - The name of the olm registry bundle image. Set to $(BUNDLE_IMG)'
	echo '    INDEX_IMG                      - The name of the olm registry index image. Set to $(INDEX_IMG)'
	echo '    DOCKER                         - Container build tool to use for building containers (e.g. podman, docker). Set to $(DOCKER)'
	echo '    GET_DIGEST_WITH                - The tool name for obtaining an image didgest. Supported tools: skopeo, podman, docker'
