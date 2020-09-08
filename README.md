# Web Terminal Operator

The Web Terminal Operator provides users with the ability to create a terminal instance embedded in the OpenShift Console.

## Deploying the operator from `next` images
After every commit in master, the index and bundle images are built and pushed to
[quay.io/repository/wto/web-terminal-operator-index:next](https://quay.io/repository/wto/web-terminal-operator-index?tab=tags) and
[quay.io/repository/wto/web-terminal-operator-metadata:next](https://quay.io/repository/wto/web-terminal-operator-metadata?tab=tags)

This repo includes a `Makefile` to simplify deploying the Operator to a cluster:

| Makefile rule | Purpose |
|---|---|
| `make install` | Register the CatalogSource and install the operator on the cluster. |
| `make register_catalogsource` | Register the CatalogSource but do not install the operator. This enables the operator to be installed manually through OperatorHub. |
| `make unregister_catalogsource` | Remove the CatalogSource from the cluster. |
| `make uninstall` | Remove the installed operator from the cluster |
| `make purge` | Like `make uninstall`, but do not fail if an error is encountered |

The commands above require being logged in to the cluster as a `cluster-admin`. See `make help` for a full list of supported environment variables and rules available.

## Deploying the operator from local sources
In order to deploy the operator from this repo directly, you need to first create the olm bundle and index, push that to a docker registry, and then create a CatalogSource referencing those images.

This can be done in one step using the Makefile `build_install` rule:
```bash
BUNDLE_IMG=#<your bundle image>
INDEX_IMG=#<your index image>
make build_install
```
This will build and push images defined by the environment variables `BUNDLE_IMG` and `INDEX_IMG`, and register a CatalogSource on the cluster. You may need to set the repos used for the index and bundle to be public to ensure they can be accessed from the cluster.

If you already have the index image pushed to your registry, then you can use the `make install` or `make register_catalogsource` rules with the environment variables defined above to install those images on the cluster.

## Removing the operator from a cluster

To remove the WebTerminal Operator and the CatalogSource use
```bash
make uninstall unregister_catalogsource
```

## Related Projects
- The Web Terminal Operator is powered by the [devworkspace-operator](https://github.com/devfile/devworkspace-operator).
