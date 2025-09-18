# Uninstalling the Web Terminal Operator
As of version 1.3.0, the Web Terminal Operator delegates creation of the cluster resources required for Web Terminals to the DevWorkspace Operator, and can be uninstalled from the OperatorHub UI normally. To uninstall the DevWorkspace Operator, manual steps are still required and should be followed according to documentation.

## Version 1.2.1 and below
Manual steps are required to uninstall the Web Terminal Operator in order to avoid issues with webhooks and finalizers used to secure terminal resources. The Web Terminal Operator utilizes validating webhooks for `pods/exec`, meaning improper uninstallation can block users' ability to exec into pods on the cluster.

1. Ensure that all DevWorkspace Custom Resources are removed along with their related k8s objects, like deployments.

	```
	kubectl delete devworkspaces.workspace.devfile.io --all-namespaces --all --wait
	kubectl delete devworkspaceroutings.controller.devfile.io --all-namespaces --all --wait
	kubectl delete components.controller.devfile.io --all-namespaces --all --wait
	```
	Note: This step must be done first, as otherwise the resources above may have finalizers that block automatic cleanup.

2. Uninstall the Operator

3. Remove the custom resource definitions installed by the operator

	```
	kubectl delete customresourcedefinitions.apiextensions.k8s.io devworkspaceroutings.controller.devfile.io
	kubectl delete customresourcedefinitions.apiextensions.k8s.io components.controller.devfile.io
	kubectl delete customresourcedefinitions.apiextensions.k8s.io devworkspaces.workspace.devfile.io
	```

4. Remove DevWorkspace Webhook Server deployment

	```
	kubectl delete deployment/devworkspace-webhook-server -n openshift-operators
	```

5. Remove lingering service, secrets, and configmaps

	```
	kubectl delete all --selector app.kubernetes.io/part-of=devworkspace-operator
	kubectl delete serviceaccounts devworkspace-webhook-server -n openshift-operators
	kubectl delete configmap devworkspace-controller -n openshift-operators
	kubectl delete clusterrole devworkspace-webhook-server
	kubectl delete clusterrolebinding devworkspace-webhook-server
	```

6. Remove mutating/validating webhook configurations.

	```
	kubectl delete mutatingwebhookconfigurations controller.devfile.io
	kubectl delete validatingwebhookconfigurations controller.devfile.io
	```