#### 1.11
- All container images used in the Web Terminal Operator have been migrated from UBI8 to UBI9 based images.
- odo has been removed from the Web Terminal tooling container image as it does not currently support UBI9 based images. See [WTO-298](https://issues.redhat.com/browse/WTO-298) for more information.
- Default tooling versions have been updated:
  - oc v4.15.0 -> v4.16.0
  - kubectl v1.28.2 -> 1.29.1
  - kustomize v5.3.0 -> v5.4.2
  - helm v3.12.1 -> v3.14.4
  - knative v1.11.2 -> v1.12.0
  - tekton v0.35.1 -> v0.37.0
  - rhoas v0.53.0 -> v0.53.0
  - submariner v0.17.0 -> v0.17.2
  - virtctl v1.2.0 -> v1.2.2

#### 1.10
- **Known issues:**
  - There is a [known issue](https://github.com/redhat-developer/web-terminal-operator/issues/162) where users logged in to an OpenShift 4.15 (or higher) cluster as `kubeadmin` are unable to create web terminals. The error message shown is `"Error Loading OpenShift command line terminal: User is not a owner of the requested workspace"`. Regular OpenShift cluster users are unaffected.
- Default tooling versions have been updated:
  - oc v4.14.5 -> v4.15.0
  - kubectl v1.27.4 -> 1.28.2
  - kustomize v5.2.1 -> v5.3.0
  - odo v3.15.0 -> v3.15.0
  - helm v3.12.1 -> v3.12.1
  - knative v1.9.2 -> v1.11.2
  - tekton v0.33.0 -> v0.35.1
  - rhoas v0.53.0 -> v0.53.0
  - submariner v0.16.2 -> v0.17.0
  - virtctl v1.1.0 -> v1.2.0

#### 1.9
- The wtoctl utility now supports adding persistent storage to existing Web Terminal instances (see `wtoctl storage --help`).
- The Web Terminal operator will update images in DevWorkspaceTemplates (`web-terminal-tooling` and `web-terminal-exec`) if they are in an unmanaged state but the image has not been changed. This allows customizing Web Terminal defaults while still receiving updated images.
- The Web Terminal tooling container now supports a wrapper for the knative CLI. If the OpenShift Serverless operator is installed in a cluster, this wrapper will prompt users to download the `kn` CLI from the Serverless operator to ensure the operator and CLI are the same version.
- Default tooling versions have been updated:
  - oc v4.13.0 -> v4.14.5
  - kubectl v1.26.1 -> 1.27.4
  - kustomize v5.0.3 -> v5.2.1
  - odo v3.9.0 -> v3.15.0
  - helm v3.11.1 -> v3.12.1
  - knative v1.7.1 -> v1.9.2
  - tekton v0.30.1 -> v0.33.0
  - rhoas v0.53.0 -> v0.53.0
  - submariner v0.14.4 -> v0.16.2
  - virtctl v0.59.0 -> v1.1.0

#### 1.8
- The wtoctl utility now supports switching between shells (see `wtoctl shell --help`). By default, the tooling image supports `bash` and `zsh`
- Default tooling versions have been updated:
  - oc v4.12.0 -> v4.13.0
  - kubectl v1.24.1 -> v1.26.1
  - kustomize v4.5.7 -> v5.0.3
  - odo v3.5.0 -> v3.9.0
  - helm v3.9.0 -> v3.11.1
  - knative v1.5.0 -> v1.7.1
  - tekton v0.24.1 -> v0.30.1
  - rhoas v0.52.0 -> v0.53.0
  - submariner v0.14.1 -> v0.14.4
  - virtctl v0.58.0 -> v0.59.0

#### 1.7
- Default tooling versions have been updated:
  - oc v4.11.2 -> v4.12.0
  - kubectl v0.24.0 -> v1.24.1
  - kustomize v4.5.7 -> v4.5.7
  - odo v2.5.1 -> v3.5.0
  - helm v3.9.0 -> v3.9.0
  - knative v1.3.1 -> v1.5.0
  - tekton v0.24.0 -> v0.24.1
  - rhoas v0.50.0 -> v0.52.0
  - submariner v0.13.0 -> v0.14.1
  - virtctl v0.56.0 -> v0.58.0

#### 1.6
- Web Terminal instances now save bash history between sessions as long as the terminal is running. History is lost when the terminal is stopped by inactivity.
- Two new CLIs are added to the default Web Terminal tooling image:
  - Kubevirt-virtctl v0.56.0
  - Kustomize v4.5.7
- Default tooling versions have been updated:
  - oc 4.10.6 -> v4.11.2
  - kubectl v0.23.0 -> v0.24.0
  - odo v2.5.0 -> v2.5.1
  - helm v3.7.1 -> v3.9.0
  - knative v1.0.0 -> v1.3.1
  - tekton v0.21.0 -> v0.24.0
  - rhoas v0.39.0 -> v0.50.0
  - submariner v0.10.1 -> v0.13.0

#### 1.5.1
- Base images for web terminal components have been updated:
  - The builder image for the web-terminal-exec container was updated to use Go 1.17
  - The Web Terminal controller now is built off of UBI 8.6

#### 1.5
- Web Terminals now show a welcome message and have a "help" command
- New `wtoctl` utility added to make configuring image and idle timeout in Web Terminal instance easier
- Added icon for Web Terminal Operator in OperatorHub UI
- Web Terminal idle timeout can now be customized more easily at the custom resource level.
- Default tooling versions have been updated:
  - oc 4.9.0 -> 4.10.6
  - odo v2.3.1 -> v2.5.0
  - helm v3.6.2 -> v3.7.1
  - knative v0.23.0 -> v1.0.0
  - tekton v0.19.1 -> v0.21.0
  - rhoas 0.34.2 -> v0.39.0
  - submariner v0.10.1 -> v0.10.1

#### 1.4
- Default tooling versions have been updated:
  - oc 4.8.2 -> 4.9.0
  - odo v2.2.3 -> v2.3.1
  - helm v3.5.0 -> v3.6.2
  - knative v0.21.0 -> v0.23.0
  - tekton v0.17.2 -> v0.19.1
  - rhoas v0.25.0 -> 0.34.2
  - submariner v0.9.1 -> v0.10.1

#### 1.3
- Default tooling versions have been updated:
  - oc 4.7.0 -> 4.8.2
  - kubectl v1.20.1 -> v0.21.0-beta.1
  - odo v2.0.4 -> v2.2.3
  - knative v0.19.1 -> v0.21.0
  - tekton 0.15.0 -> 0.17.2
  - kubectx & kubectx v0.9.2 -> v0.9.4
  - rhoas v0.24.1 -> v0.25.0
  - submariner 0.9.1 (first release)

#### 1.2

- Cluster administrators can now access the terminal on OpenShift 4.7 and up
- A message is displayed when the terminal has been idled
- Default tooling versions have been updated:
  - oc 4.6.1 -> 4.7.0
  - kubectl v1.19.0 -> v1.20.1
  - odo v2.0.0 -> v2.0.4
  - helm v3.3.4 -> v3.5.0
  - knative v0.16.1 -> v0.19.1
  - tekton 0.11.0 -> 0.15.0
  - kubectx & kubectx v0.9.1 -> v0.9.2

#### 1.1

- Default tooling versions have been updated:
  - oc 4.5.3 -> 4.6.1
  - kubectl v1.18.2-0-g52c56ce -> v1.19.0
  - odo v1.2.4 -> v2.0.0
  - helm v3.2.3 -> v3.3.4
  - knative v0.13.2 -> v0.16.1
  - tekton 0.9.0 -> 0.11.0

#### 1.0.1

- Initial release of the operator
