SHELL := bash
.SHELLFLAGS = -ec

BUNDLE_IMG ?= quay.io/wto/web-terminal-operator-metadata:next
INDEX_IMG ?= quay.io/wto/web-terminal-operator-index:next
LATEST_INDEX_IMG ?= quay.io/wto/web-terminal-operator-index:v1.0.1

.ONESHELL:
all: help

_print_vars:
	@echo "Current env vars:"
	echo "    BUNDLE_IMG=$(BUNDLE_IMG)"
	echo "    INDEX_IMG=$(INDEX_IMG)"
	echo "    LATEST_INDEX_IMG=$(LATEST_INDEX_IMG)"

### update_dependencies: updates files from DevWorkspace API and Operators
update_dependencies:
	./update-dependencies.sh

### gen_terminal_csv: generate the csv for a newer version. Refer to gen_terminal_csv makefile definition for extra manual steps that are needed.
gen_terminal_csv : update_dependencies
	# Remove everything in manifests except CSV which contains manually filled
	# fields, like description, puslisher, ...
	find ./manifests -type f -not -name 'web-terminal.clusterserviceversion.yaml' -delete

	# Need to be in root of the controller in order to run operator-sdk
	pushd devworkspace-dependencies > /dev/null
	operator-sdk generate csv --apis-dir ./pkg/apis --csv-version 1.0.0 --make-manifests --update-crds --operator-name "web-terminal" --output-dir ../
	popd > /dev/null

### build: build the terminal bundle and index and push them to a docker registry
build: _print_vars _check_imgs_env _check_skopeo_installed
	@rm -rf ./generated
	# Create the bundle and push it to a docker registry
	operator-sdk bundle create $(BUNDLE_IMG) --channels alpha --package web-terminal --directory ./manifests --overwrite --output-dir generated
	docker push $(BUNDLE_IMG)

	BUNDLE_DIGEST=$$(skopeo inspect docker://$(BUNDLE_IMG) | jq -r '.Digest')
	BUNDLE_IMG=$(BUNDLE_IMG)
	BUNDLE_IMG_DIGEST="$${BUNDLE_IMG%:*}@$${BUNDLE_DIGEST}"
	# create / update and push an index that contains the bundle
	opm index add -c docker --bundles $${BUNDLE_IMG_DIGEST} --tag $(INDEX_IMG) --from-index $(LATEST_INDEX_IMG)
	docker push $(INDEX_IMG)

### export: export the bundles stored in the index to the exported-manifests folder
export: _print_vars _check_imgs_env
	@rm -rf ./exported-manifests
	# Export the bundles with the name web-terminal inside of $(INDEX_IMG)
	# This command basic exports the index back into the old format
	opm index export -c docker -f exported-manifests -i $(INDEX_IMG) -o web-terminal

### register_catalogsource: creates the catalogsource to make the operator be available on the marketplace. Image referenced by INDEX_IMG must be pushed and publicly available
register_catalogsource: _print_vars _check_imgs_env _check_skopeo_installed
	@INDEX_DIGEST=$$(skopeo inspect docker://$(INDEX_IMG) | jq -r '.Digest')
	INDEX_IMG=$(INDEX_IMG)
	INDEX_IMG_DIGEST="$${INDEX_IMG%:*}@$${INDEX_DIGEST}"

	# replace references of catalogsource img with your image
	sed -i.bak -e "s|quay.io/che-incubator/che-workspace-operator-index:latest|$${INDEX_IMG_DIGEST}|g" ./catalog-source.yaml
	# use ';' to make sure we undo changes to catalog-source.yaml even if command fails.
	oc apply -f ./catalog-source.yaml ; \
	  mv ./catalog-source.yaml.bak ./catalog-source.yaml

### unregister_catalogsource: remove the catalogsource from the cluster.
unregister_catalogsource: _print_vars
	@oc delete -f ./catalog-source.yaml --ignore-not-found=true

	oc apply -f ./mirror-index-manifests/imageContentSourcePolicy.yaml

### unregister_catalogsource: unregister the catalogsource and delete the imageContentSourcePolicy
unregister_catalogsource:
	oc delete catalogsource custom-web-terminal-catalog -n openshift-marketplace --ignore-not-found
	oc delete imagecontentsourcepolicy web-terminal-index-mirror --ignore-not-found

### build_install: build the catalog and create catalogsource and operator subscription on the cluster
build_install: _print_vars build install

### install: creates catalog source along with operator subscription on the cluster
install: _print_vars register_catalogsource
	oc apply -f ./operator-subscription.yaml

### uninstall: uninstalls the Web Terminal Operator in the proper way described in documentation
uninstall:
	# 1. Ensure that all DevWorkspace Custom Resources are removed to avoid issues with finalizers
	# make sure depending objects are clean up as well
	kubectl delete devworkspaces.workspace.devfile.io --all-namespaces --all --wait
	kubectl delete workspaceroutings.controller.devfile.io --all-namespaces --all --wait
	kubectl delete components.controller.devfile.io --all-namespaces --all --wait
	# 2. Uninstall the Operator
	kubectl delete subscriptions.operators.coreos.com web-terminal -n openshift-operators
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
	# Extra step that is not described in admin guide
	kubectl delete csv web-terminal.v1.0.0 -n openshift-operators

### purge: uninstalls the Web Terminal Operator in the safe way where every object is removed without failing when not found. Use it if uninstall fails because of no objects found
purge:
	# 1. Ensure that all DevWorkspace Custom Resources are removed to avoid issues with finalizers
	# make sure depending objects are clean up as well
	if [ ! -z $(kubectl get customresourcedefinitions.apiextensions.k8s.io devworkspaces.workspace.devfile.io --ignore-not-found) ]
	then
		kubectl delete devworkspaces.workspace.devfile.io --all-namespaces --all --wait
	fi
	if [ ! -z $(kubectl get customresourcedefinitions.apiextensions.k8s.io workspaceroutings.controller.devfile.io --ignore-not-found) ]
	then
		kubectl delete workspaceroutings.controller.devfile.io --all-namespaces --all --wait
	fi
	if [ ! -z $(kubectl get customresourcedefinitions.apiextensions.k8s.io components.controller.devfile.io --ignore-not-found) ]
	then
		kubectl delete components.controller.devfile.io --all-namespaces --all --wait
	fi
	# 2. Uninstall the Operator
	kubectl delete subscriptions.operators.coreos.com web-terminal -n openshift-operators --ignore-not-found
	# 3. Remove CRDs
	kubectl delete customresourcedefinitions.apiextensions.k8s.io workspaceroutings.controller.devfile.io --ignore-not-found
	kubectl delete customresourcedefinitions.apiextensions.k8s.io components.controller.devfile.io --ignore-not-found
	kubectl delete customresourcedefinitions.apiextensions.k8s.io devworkspaces.workspace.devfile.io --ignore-not-found
	# 4. Remove DevWorkspace Webhook Server Deployment itself
	kubectl delete deployment/devworkspace-webhook-server -n openshift-operators --ignore-not-found
	# 5. Remove lingering service, secrets, and configmaps
	kubectl delete all --selector app.kubernetes.io/part-of=devworkspace-operator,app.kubernetes.io/name=devworkspace-webhook-server
	kubectl delete serviceaccounts devworkspace-webhook-server -n openshift-operators --ignore-not-found
	kubectl delete configmap devworkspace-controller -n openshift-operators --ignore-not-found
	kubectl delete clusterrole devworkspace-webhook-server --ignore-not-found
	kubectl delete clusterrolebinding devworkspace-webhook-server --ignore-not-found
	# 6. Remove mutating/validating webhooks configuration
	kubectl delete mutatingwebhookconfigurations controller.devfile.io --ignore-not-found
	kubectl delete validatingwebhookconfigurations controller.devfile.io --ignore-not-found
	# Extra step that is not described in admin guide
	kubectl delete csv web-terminal.v1.0.0 -n openshift-operators --ignore-not-found

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
	@echo 'Available rules:'
	sed -n 's/^### /    /p' $< | awk 'BEGIN { FS=":" } { printf "%-34s -%s\n", $$1, $$2 }'
	echo ''
	echo 'Supported environment variables:'
	echo '    BUNDLE_IMG                     - The name of the olm registry bundle image. Set to $(BUNDLE_IMG)'
	echo '    INDEX_IMG                      - The name of the olm registry index image. Set to $(INDEX_IMG)'
	echo '    LATEST_INDEX_IMG               - The name of the last versions index. Set to $(LATEST_INDEX_IMG)'
	echo '    DEVWORKSPACE_API_VERSION       - Branch or tag of the github.com/devfile/kubernetes-api to depend on.'
	echo '    DEVWORKSPACE_OPERATOR_VERSION  - The branch/tag of the terminal manifests.'
