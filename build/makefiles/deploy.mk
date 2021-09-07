
# it's easier to bump whole kubeconfig instead of grabbing cluster URL from the current context
_bump_kubeconfig:
	mkdir -p $(INTERNAL_TMP_DIR)
ifndef KUBECONFIG
	$(eval CONFIG_FILE = ${HOME}/.kube/config)
else
	$(eval CONFIG_FILE = ${KUBECONFIG})
endif
	cp $(CONFIG_FILE) $(BUMPED_KUBECONFIG)

_login_with_devworkspace_sa:
	$(eval SA_TOKEN := $(shell $(K8S_CLI) get secrets -o=json -n $(NAMESPACE) | jq -r '[.items[] | select (.type == "kubernetes.io/service-account-token" and .metadata.annotations."kubernetes.io/service-account.name" == "$(DEVWORKSPACE_CTRL_SA)")][0].data.token' | base64 --decode ))
	echo "Logging as controller's SA in $(NAMESPACE)"
	oc login --token=$(SA_TOKEN) --kubeconfig=$(BUMPED_KUBECONFIG)

### debug: Runs the controller locally with debugging enabled, watching cluster defined in ~/.kube/config
debug:
	MAX_CONCURRENT_RECONCILES="5" \
	CONTROLLER_SERVICE_ACCOUNT_NAME="web-terminal-controller" \
    	WATCH_NAMESPACE="openshift-operators" \
    	dlv debug --listen=:2345 --headless=true --api-version=2 ./main.go --