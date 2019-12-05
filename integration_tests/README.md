## How to run the tests

The following environment variables are needed:

|Name|Description|Default|
|----|-----------|----------|
|KUBECONFIG |File for a cluster to use for tests. Tests can create and delete namespaces | `$HOME/.kube/config`|
|OPERATOR_IMAGE |Docker image of the operator to test |`armory/spinnaker-operator:dev`|
|HALYARD_IMAGE |Docker image of Halyard to use for tests |`armory/halyard:operator-0.3.x`|
|CLEAN_START |Indicates if all namespaces used by the tests should be deleted *before* tests start if they exist (can greatly increase execution time)|`false`|

Run tests with `make integration_test`
