# Web Terminal Operator

Web Terminal Operator provides an ability for users to use terminal embedded into OpenShift Console.

### Deploying the controller with olm
In order to deploy the operator you need to first create the olm bundle, olm index, push that to a docker registry then create catalog source with and install WebTerminal Operator on your cluster.
You can do this by using the Makefile build_install rule:
```bash
make build_install
```
Before doing this you need to set environment variables BUNDLE_IMG, INDEX_IMG

When this is done running you'll need to go to your docker registry and make the created repos public (they are private by default)

If you already have the index image on your docker registry then you can use the rule
```bash
make install
```
to install WebTerminal Operator on your cluster.

There is also an ability to register CatalogSource with WebTerminal Operator without installing it, to do it:
```bash
make register_catalogsource
```
After that admins are able to set WebTerminal Operator on OperatorHub and install it when needed.

To remove the WebTerminal Operator along with CatalogSource use
```bash
make uninstall
```

### Related Projects
- The Web Terminal Operator is powered by the [devworkspace-operator](https://github.com/devfile/devworkspace-operator)
