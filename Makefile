DEVWORKSPACE_API_VERSION ?= master
DEVWORKSPACE_OPERATOR_VERSION ?= master
BUNDLE_IMG ?=
INDEX_IMG ?=

.ONESHELL:
all: help

_print_vars:
	@echo "Current env vars:"
	@echo "    DEVWORKSPACE_API_VERSION=$(DEVWORKSPACE_API_VERSION)"
	@echo "    DEVWORKSPACE_OPERATOR_VERSION=$(DEVWORKSPACE_OPERATOR_VERSION)"
	@echo "    BUNDLE_IMG=$(BUNDLE_IMG)"
	@echo "    INDEX_IMG=$(INDEX_IMG)"

update_dependencies:
	./collect-sources.sh

### gen_terminal_csv : generate the csv for a newer version. Refer to gen_terminal_csv makefile definition for extra manual steps that are needed.
gen_terminal_csv : update_dependencies
	# Some steps need to be done manually in order to get the csv ready
	# This includes:
	# 1. Updating the description
	# 2. Remove the edit and view workspaces role from the csv
	# 3. Update the alm-examples (they are reset on each csv generate)

	# Need to be in root of the controller in order to run operator-sdk
	pushd devworkspace-dependencies > /dev/null
	operator-sdk generate csv --apis-dir ./pkg/apis --csv-version 1.0.0 --make-manifests --update-crds --operator-name "web-terminal" --output-dir ../
	popd > /dev/null

	# Add in the edit workspaces and view workspaces cluster roles
	cp devworkspace-operator/deploy/edit-workspaces-cluster-role.yaml manifests/
	cp devworkspace-operator/deploy/view-workspaces-cluster-role.yaml manifests/

### olm_build_bundle_index: build the terminal bundle and index and push them to a docker registry
olm_build_bundle_index: _print_vars _check_imgs_env
	# Create the bundle and push it to a docker registry
	operator-sdk bundle create $(BUNDLE_IMG) --channels alpha --package web-terminal --directory ./manifests --overwrite --output-dir generated
	docker push $(BUNDLE_IMG)

	# create / update and push an index that contains the bundle
	opm index add -c docker --bundles $(BUNDLE_IMG) --tag $(INDEX_IMG)
	docker push $(INDEX_IMG)

### olm_install_local: use the catalogsource to make the operator be available on the marketplace. Must have $(INDEX_IMG) available on docker registry already and have it set to public
olm_install_local: _print_vars _check_imgs_env
	# replace references of catalogsource img with your image
	sed -i.bak -e  "s|quay.io/che-incubator/che-workspace-operator-index:latest|$(INDEX_IMG)|g" ./catalog-source.yaml
	oc apply -f ./catalog-source.yaml
	sed -i.bak -e "s|$(INDEX_IMG)|quay.io/che-incubator/che-workspace-operator-index:latest|g" ./catalog-source.yaml

	# remove the .bak files
	rm ./catalog-source.yaml.bak

### olm_build_install_local: build the catalog and deploys the catalog to the cluster
olm_build_install_local: _print_vars olm_build_bundle_index olm_install_local

### olm_uninstall: uninstalls the operator
olm_uninstall:
	oc delete catalogsource web-terminal-crd-registry -n openshift-marketplace

_check_imgs_env:
ifndef BUNDLE_IMG
	$(error "BUNDLE_IMG not set")
endif
ifndef INDEX_IMG
	$(error "INDEX_IMG not set")
endif

.PHONY: help
### help: print this message
help: Makefile
	@echo 'Available rules:'
	@sed -n 's/^### /    /p' $< | awk 'BEGIN { FS=":" } { printf "%-32s -%s\n", $$1, $$2 }'
	@echo ''
	@echo 'Supported environment variables:'
	@echo '    DEVWORKSPACE_API_VERSION       - Branch or tag of the github.com/devfile/kubernetes-api to depend on. Defaults to master'
	@echo '    DEVWORKSPACE_OPERATOR_VERSION  - The branch/tag of the terminal manifests'
	@echo '    BUNDLE_IMG                     - The name of the olm registry bundle image'
	@echo '    INDEX_IMG                      - The name of the olm registry index image'
