## How to run the tests

The following environment variables are needed:

|Name|Description|Default|
|----|-----------|----------|
|KUBECONFIG |File for a cluster to use for tests. Tests can create and delete namespaces. | `$HOME/.kube/config`|
|OPERATOR_IMAGE |Docker image of the operator to test |`armory/spinnaker-operator:dev`|
|HALYARD_IMAGE |Docker image of Halyard to use for tests |`armory/halyard:operator-0.3.x`|
|S3_BUCKET |S3 bucket name used for spinnaker persistence. Worker nodes should have access to it. |`operator-int-tests`|
|S3_BUCKET_REGION |Region used by the S3 bucket. |`us-west-2`|

Run tests with `make integration_test`.

Tests create dynamic namespaces, they are not deleted after finishing tests. There are two reasons for this:
1. Deleting a namespace can take a lot of time, and usually this can be done as part of regular maintenance tasks outside of test execution.
2. If a test fails, it's useful to have the namespaces available to be able to inspect the state of the operator and spinnaker, like detailed log files.
