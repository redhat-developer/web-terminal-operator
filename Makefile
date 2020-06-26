DEVWORKSPACE_API_VERSION ?= master
DEVWORKSPACE_OPERATOR_VERSION ?= master
BUNDLE_IMG ?= ""
INDEX_IMG ?= ""

all: help

_print_vars:
	@echo "Current env vars:"
	@echo "    DEVWORKSPACE_API_VERSION=$(DEVWORKSPACE_API_VERSION)"
	@echo "    DEVWORKSPACE_OPERATOR_VERSION=$(DEVWORKSPACE_OPERATOR_VERSION)"
	@echo "    BUNDLE_IMG=$(BUNDLE_IMG)"
	@echo "    INDEX_IMG=$(INDEX_IMG)"

### update_devworkspace_crds: pull latest devworkspace CRDs to ./devworkspace-crds. Pulls $(DEVWORKSPACE_API_VERSION) branch/tag
.ONESHELL:
update_devworkspace_crds:
	@mkdir -p devworkspace-crds
	cd devworkspace-crds
	if [ ! -d ./.git ]; then
		git init
		git remote add origin -f https://github.com/devfile/kubernetes-api.git
		git config core.sparsecheckout true
		echo "deploy/crds/*" > .git/info/sparse-checkout
	else
		git remote set-url origin https://github.com/devfile/kubernetes-api.git
	fi
	git fetch --tags -p origin
	if git show-ref --verify refs/tags/$(DEVWORKSPACE_API_VERSION) --quiet; then
		echo 'DevWorkspace API is specified from tag'
		git checkout tags/$(DEVWORKSPACE_API_VERSION)
	else
		echo 'DevWorkspace API is specified from branch'
		git checkout $(DEVWORKSPACE_API_VERSION) && git reset --hard origin/$(DEVWORKSPACE_API_VERSION)
	fi

### update_devworkspace_operator: pull latest operator to ./devworkspace-operator. Pulls $(DEVWORKSPACE_OPERATOR_VERSION) branch/tag
.ONESHELL:
update_devworkspace_operator:
	@mkdir -p devworkspace-operator
	cd devworkspace-operator
	if [ ! -d ./.git ]; then
		git init
		git remote add origin -f https://github.com/devfile/devworkspace-operator.git
	else
		git remote set-url origin https://github.com/devfile/devworkspace-operator.git
	fi
	git fetch --tags -p origin
	if git show-ref --verify refs/tags/$(DEVWORKSPACE_OPERATOR_VERSION) --quiet; then
		echo 'Operator is specified from tag'
		git checkout tags/$(DEVWORKSPACE_OPERATOR_VERSION)
	else
		echo 'Operator is specified from branch'
		git checkout $(DEVWORKSPACE_OPERATOR_VERSION) && git reset --hard origin/$(DEVWORKSPACE_OPERATOR_VERSION)
	fi

### gen_terminal_csv : generate the csv for a newer version
gen_terminal_csv : update_devworkspace_crds update_devworkspace_operator
	# Need to be in root of the controller in order to run operator-sdk
	cd devworkspace-operator
	operator-sdk generate csv --apis-dir ./devworkspace-operator/pkg/apis --csv-version 1.0.0 --make-manifests --update-crds --operator-name "web-terminal" --output-dir ../
	
	# filter the deployments so that only the valid deployment is available. See: https://github.com/eclipse/che/issues/17010
	cat ../manifests/web-terminal.clusterserviceversion.yaml | \
	yq -Y \
	'.spec.install.spec.deployments[] |= select( .spec.selector.matchLabels.app? and .spec.selector.matchLabels.app=="che-workspace-controller")' | \
	tee ../manifests/web\ terminal.clusterserviceversion.yaml >>/dev/null

	cp ../devworkspace-crds/deploy/crds/workspace.devfile.io_devworkspaces_crd.yaml ./manifests
	
	# Add in the edit workspaces and view workspaces cluster roles
	cp ./deploy/edit-workspaces-cluster-role.yaml ./manifests
	cp ./deploy/view-workspaces-cluster-role.yaml ./manifests

### olm_build_bundle_index: build the terminal bundle and index and push them to a docker registry
olm_build_bundle_index: _print_vars check-env
	# Create the bundle and push it to a docker registry
	operator-sdk bundle create $(BUNDLE_IMG) --channels alpha --package web-terminal --directory ./manifests --overwrite --output-dir generated
	docker push $(BUNDLE_IMG)

	# create / update and push an index that contains the bundle
	opm index add -c docker --bundles $(BUNDLE_IMG) --tag $(INDEX_IMG)
	docker push $(INDEX_IMG)

### olm_install_local: use the catalogsource to make the operator be available on the marketplace. Must have $(INDEX_IMG) available on docker registry already and have it set to public
olm_install_local: _print_vars
	# replace references of catalogsource img with your image
	sed -i.bak -e  "s|quay.io/che-incubator/che-workspace-operator-index:latest|$(INDEX_IMG)|g" ./catalog-source.yaml
	oc apply -f ./catalog-source.yaml
	sed -i.bak -e "s|$(INDEX_IMG)|quay.io/che-incubator/che-workspace-operator-index:latest|g" ./catalog-source.yaml

	# remove the .bak files
	rm ./catalog-source.yaml.bak

### olm_build_install_local: build the catalog and deploys the catalog to the cluster
olm_build_install_local: _print_vars olm_build_bundle olm_create_index olm_start_local

### olm_uninstall: uninstalls the operator
olm_uninstall:
	oc delete catalogsource che-workspace-crd-registry -n openshift-marketplace

check-env:
	if test "$(BUNDLE_IMG)" = "" ; then \
		echo "BUNDLE_IMG not set"; \
		exit 1; \
	fi
	if test "$(INDEX_IMG)" = "" ; then \
		echo "INDEX_IMG not set"; \
		exit 1; \
	fi

.PHONY: help
### help: print this message
help: Makefile
	@echo 'Available rules:'
	@sed -n 's/^### /    /p' $< | awk 'BEGIN { FS=":" } { printf "%-30s -%s\n", $$1, $$2 }'
	@echo ''
	@echo 'Supported environment variables:'
	@echo '    DEVWORKSPACE_API_VERSION       - Branch or tag of the github.com/devfile/kubernetes-api to depend on. Defaults to master'
	@echo '    DEVWORKSPACE_OPERATOR_VERSION  - The branch/tag of the terminal manifests'
	@echo '    BUNDLE_IMG                     - The name of the olm registry bundle image'
	@echo '    INDEX_IMG                      - The name of the olm registry index image'
