## How to run the tests

The following environment variables are needed:

|Name|Description|Default|
|----|-----------|----------|
|KUBECONFIG |File for a cluster to use for tests. Tests can create and delete namespaces. | `$HOME/.kube/config`|
|OPERATOR_IMAGE |Docker image of the operator to test |`armory/spinnaker-operator:dev`|
|HALYARD_IMAGE |Docker image of Halyard to use for tests |`armory/halyard:operator-0.3.x`|
|S3_BUCKET |S3 bucket name used for spinnaker persistence. Worker nodes should have access to it. |`spin-int-tests`|

Run tests with `make integration_test`
