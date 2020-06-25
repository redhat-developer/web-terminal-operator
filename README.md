# Web Terminal Operator

Web Terminal Operator provides an ability for users to use terminal embedded into OpenShift Console.

### Deploying the controller with olm
In order to deploy the operator you need to first create the olm bundle, olm index, push that to quay then deploy the crd registry to your cluster.
You can do this by using the Makefile olm_build_start_terminal_local rule:
```bash
make olm_build_start_terminal_local
```
Before doing this you need to set environment variables BUNDLE_IMG, INDEX_IMG

When this is done running you'll need to go to your docker registry and make the created repos public (they are private by default)

If you already have the index image on your docker registry then you can use the rule
```bash
make olm_start_terminal_local
```
to deploy the crd registry to your cluster

To remove the crd registry use
```bash
make olm_terminal_uninstall
```

