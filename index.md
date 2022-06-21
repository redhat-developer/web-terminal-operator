# Table of Contents

- [Introduction](#introduction)
- [How to use](#how-to-use)
- [Installation](#installation)
- [Uninstalling](#uninstalling)

# Introduction

The Web Terminal Operator allows you to start a terminal inside of your browser with common CLI tools for interacting with the cluster.

**Note:** The OpenShift console integration that allows easily creating web terminal instances
and logging in automatically is available in OpenShift 4.5.3 and higher. In previous versions of
OpenShift, the operator can be installed but web terminals will have to be created and accessed
manually.

![image](./images/introduction.png)

# <a id="how-to-use"></a>How to use

After installing the Web Terminal operator, you can use the web terminal by first clicking on the terminal button in the top right of the OpenShift console

![image](./images/masthead-icon.png)

This will open up a web terminal at the bottom of your screen. This terminal will automatically be logged in as your OpenShift user and have tools like `oc`, `kubectl`, `odo`, Knative, Tekton, Helm, `kubens`, and `kubectx` pre-installed.

Once you are logged into the terminal, you can type the `help` command to see a list of installed CLI tools. The tooling image also comes with a utility named `wtoctl` to aid in customizing the running web terminal. See `wtoctl help` for available commands.

![image](./images/initialization.gif)

# Installation

The Web Terminal can be installed via OperatorHub on Openshift Clusters. To install, press the **Install** button, choose the upgrade strategy, and wait for the **Installed** Operator status.

![image](./images/installation.png)

When the operator is installed, and you refresh your page, you will see a terminal button appear on the top right of the console.

![image](./images/masthead-icon.png)

# Uninstalling

Parts of the operator must be manually uninstalled for security purposes. It also allows you to save cluster resources, as terminals cannot be idled when the operator is uninstalled. In order to fully uninstall an admin must also remove the DevWorkspace Operator, which is installed with the Web Terminal Operator as a dependency.

## Removing the Web Terminal Operator
1. Uninstall the Web Terminal Operator using the web console:
    1. In the *Administrator* perspective of the web console, navigate to *Operators -> Installed Operators*.
    2. Scroll the filter list or type a keyword into the *Filter by name* box to find the *Web Terminal* Operator.
    3. Click the Options menu for the Web Terminal Operator, and then select *Uninstall Operator*.
    4. In the *Uninstall Operator* confirmation dialog box, click *Uninstall* to remove the Operator, Operator deployments, and pods from the cluster. The Operator stops running and no longer receives updates.

2. Remove all Web Terminal custom resouces from the cluster.
    ```bash
    kubectl delete devworkspaces.workspace.devfile.io --all-namespaces \
        --selector 'console.openshift.io/terminal=true' --wait
    kubectl delete devworkspacetemplates.workspace.devfile.io --all-namespaces \
        --selector 'console.openshift.io/terminal=true' --wait
    ```

## Removing the DevWorkspace Operator dependency
1. Ensure that all DevWorkspace Custom Resources are removed along with their related k8s objects, like deployments. It is crucial that this is done first, otherwise finalizers might make it difficult to fully uninstall the operator.
    ```
    kubectl delete devworkspaces.workspace.devfile.io --all-namespaces --all --wait
    kubectl delete devworkspaceroutings.controller.devfile.io --all-namespaces --all --wait
    ```

2. Remove the custom resource definitions
    ```
    kubectl delete customresourcedefinitions.apiextensions.k8s.io devworkspaceroutings.controller.devfile.io
    kubectl delete customresourcedefinitions.apiextensions.k8s.io devworkspaces.workspace.devfile.io
    kubectl delete customresourcedefinitions.apiextensions.k8s.io devworkspacetemplates.workspace.devfile.io
    kubectl delete customresourcedefinitions.apiextensions.k8s.io devworkspaceoperatorconfigs.controller.devfile.io
    ```

3. Verify that all involved custom resource definitions are removed. The following command should not display any result.
    ```bash
    kubectl get customresourcedefinitions.apiextensions.k8s.io | grep "devfile.io"
    ```

4. Remove the `devworkspace-webhook-server` deployment along with mutating and validating webhooks:
    ```
    kubectl delete deployment/devworkspace-webhook-server -n openshift-operators
    kubectl delete mutatingwebhookconfigurations controller.devfile.io
    kubectl delete validatingwebhookconfigurations controller.devfile.io
    ```

5. Remove any remaining services, secrets, and config maps. Depending on the installation, some resources included in the following command may not exist on the cluster.

    ```bash
    kubectl delete all --selector 'app.kubernetes.io/part-of=devworkspace-operator,app.kubernetes.io/name=devworkspace-webhook-server'
    kubectl delete serviceaccounts devworkspace-webhook-server -n openshift-operators
    kubectl delete configmap devworkspace-controller -n openshift-operators
    kubectl delete clusterrole devworkspace-webhook-server
    kubectl delete clusterrolebinding devworkspace-webhook-server
    ```

6. Remove the Web Terminal Operator via the OperatorHub UI
    1. In the *Administrator* perspective of the web console, navigate to *Operators -> Installed Operators*.
    2. Scroll the filter list or type a keyword into the *Filter by name* box to find the *DevWorkspace* Operator.
    3. Click the Options menu for the DevWorkspace Operator, and then select *Uninstall Operator*.
    4. In the *Uninstall Operator* confirmation dialog box, click *Uninstall* to remove the Operator, Operator deployments, and pods from the cluster. The Operator stops running and no longer receives updates.
