# Spinnaker Operator for Kubernetes

A Kubernetes native application can roughly be defined by two traits. First, it is deployed to Kubernetes. Second, it is managed using the Kubernetes toolchain (APIs, manifests, kubectl, etc.) A Kubernetes Operator is the runtime that manages a Kubernetes native application. As a concept, an operator can be thought of as the codification of the operating practices normally conducted by a person.

The sophistication of an operator can vary. However, the initial scope of an operator is often to manage the installation of the application. More advanced operator scope usually revolves around seamless upgrades and automatic failure recovery.

## Implementation

This operator acts as a layer on top of Halyard. Under the hood it runs Halyard commands via Kubernetes Job objects.

## Operator Installation

To install the operator run:

```
$ git clone https://github.com/armory-io/spinnaker-operator
$ cd spinnaker-operator
$ kubectl apply -f deploy/
```

This will add a custom resource called `SpinnakerService`,  a deployment of the operator, and a service account with the needed permissions.

## Usage

### Custom Resources

You can manage a Spinnaker installation with `kubectl`. The `SpinnakerService` objects will now be available with the normal `kubectl` verbs.

Such as,

Get:
```
$ kubectl get spinnakerservice
NAME                         AGE
spinnakerinstallation-v001   31m
```

Delete:
```
$ kubectl delete spinnakerservice spinnakerinstallation-v001
spinnakerservice.spinnaker.armory.io "spinnakerinstallation-v001" deleted
```

The following is an example of a `SpinnakerService` manifest,

```
apiVersion: spinnaker.armory.io/v1alpha1
kind: SpinnakerService
metadata:
  name: spinnakerinstallation-v001
spec:
  halConfigMap: halconfig-v001
```


### Operator Deployment

Once installed you should see a new deployment representing the operator. The operator listens to the Kubernetes event stream for changes to the `SpinnakerService` objects. You can check on the status of the operator using `kubectl`. For example,

```
$ kubectl get deployments
NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
spinnaker-operator   1         1         1            1           38m
```

### Spinnaker Installation/Upgrade via the Operator

Two manifests are needed to install or upgrade Spinnaker. The halconfig should be put into a config map. Then a custom resource definition with kind `SpinnakerService` should reference the config map.

An example can be found in `deploy/example`. If you would like to try it out you can run:
```
$ kubectl apply -f deploy/example/
```
*Note:* The example halconfig provided has some environment specific information. Such as what S3 bucket to use for storage and what namespace to install Spinnaker. You will probably want to look through it and change values that make sense for your environment.
