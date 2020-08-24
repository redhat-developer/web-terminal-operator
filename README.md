# Web Terminal Operator

Web Terminal Operator provides an ability for users to use terminal embedded into OpenShift Console.

### Deploying next operator from next images
After every commit in master index and bundle images are built and pushed to:
[quay.io/repository/wto/web-terminal-operator-index:next](https://quay.io/repository/wto/web-terminal-operator-index?tab=tags)
[quay.io/repository/wto/web-terminal-operator-metadata:next](https://quay.io/repository/wto/web-terminal-operator-metadata?tab=tags)

To try them you can just do
```
make install
```
and wait until Operator is installed on the cluster.

There is also an ability to register CatalogSource with WebTerminal Operator without installing it, to do it:
```bash
make register_catalogsource
```
After that admins are able to set WebTerminal Operator on OperatorHub and install it when needed.

### Deploying the operator from local sources
In order to deploy the operator you need to first create the olm bundle, olm index, push that to a docker registry then create catalog source with and install WebTerminal Operator on your cluster.
You can do this by using the Makefile build_install rule:
```bash
make build_install
```
Before doing this you need to set environment variables BUNDLE_IMG, INDEX_IMG

When this is done running you'll need to go to your docker registry and make the created repos public (they are private by default)

If you already have the index image on your docker registry then you can use `make install` or `make register_catalogsource` makefile rules.

### Removing operator

To remove the WebTerminal Operator along with CatalogSource use
```bash
make uninstall
```

### Related Projects
- The Web Terminal Operator is powered by the [devworkspace-operator](https://github.com/devfile/devworkspace-operator)
