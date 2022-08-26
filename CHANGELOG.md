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
