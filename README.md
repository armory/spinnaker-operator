# Spinnaker Operator for Kubernetes

We've announced the [Spinnaker Operator](https://blog.armory.io/spinnaker-operator/): a Kubernetes operator to deploy and manage Spinnaker with the tools you're used to. We're sharing configuration in this repository (code to come soon) to let the community evaluate it and provide feedback.

## Goals
The Spinnaker operator:
- should be able to install any version of Spinnaker with a published BOM
- should perform preflight checks to confidently upgrade Spinnaker

More concretely, the operator:
- is configured via a `configMap` or a `secret`
- can deploy in a single namespace or in multiple namespaces
- garbage collect configuration (secrets, deployments, ...)
- provides a validating admission webhook to validate the configuration before it is applied

We plan to support many validations such as provider (AWS, Kubernetes,...) validation, connectivity to CI. Please let us know what would make your life easier when installing Spinnaker! You can use GitHub issues for the time being.


## Limitations
*The operator is in alpha and its CRD may change quite a bit. It is actively being developed.*
- Spinnaker configuration in `secret` is not supported at the moment.

## Requirements
The validating admission controller [requires](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites):
- Kubernetes server v1.13+
- Admission controllers enabled (`-enable-admission-plugins`)
- `ValidatingAdmissionWebhook` enabled in the kube-apiserver (should be the default)

Note: If you can't use the validation webhook, pass the `--without-admission-controller` to the operator (like in `deploy/operator/basic/deployment.yaml`).

## Operator Installation

First we'll install the `SpinnakerService` CRD:

```bash
$ git clone https://github.com/armory/spinnaker-operator
$ cd spinnaker-operator
$ kubectl apply -f deploy/crds/spinnaker_v1alpha1_spinnakerservice_crd.yaml
```

There are two modes for the operator:
- a basic mode that can install Spinnaker in a single namespace. In this mode, there's no validating admission webhook.
- a cluster mode that requires a `ClusterRole` to perform validation.

The main difference between the two modes is that basic only requires a `Role` (vs a `ClusterRole`) and has no validating webhook.

Once installed you should see a new deployment representing the operator. The operator watches for changes to the `SpinnakerService` objects. You can check on the status of the operator using `kubectl`.

### Basic install (no validating webhook)
To install the operator run:

```bash
$ kubectl apply -n <namespace> -f deploy/operator/basic
```

`namespace` is the namespace where you want the operator to live and deploy to.

### Cluster install
To install the operator run:

```bash
$ kubectl apply -n <namespace> -f deploy/operator/cluster
```

`namespace` is the namespace where you want the operator to live.


## Spinnaker Installation

Once you've installed the operator, you can install Spinnaker by making a configuration (`configMap`). Check out examples in `deploy/spinnaker/examples`. If you prefer to use `kustomize`, we've added some kustomization in `deploy/spinnaker/kustomize` (WIP)


### Example 1: Installing version 1.15.1

```bash
$ kubectl -n <namespace> apply -f deploy/spinnaker/examples/basic
```

This configuration does not contain any connected accounts, just a persistent storage. 

### Example 2: Using Kustomize (TODO)

Set your own values in `deploy/spinnaker/kustomize/kustomization.yaml`, then:
 

```bash
$ kustomize build deploy/spinnaker/kustomize | kubectl -n <namespace> apply -f -
```

Or if using `kubectl` version 1.14+:
```bash
$ kubectl -n <namespace> apply -f deploy/spinnaker/examples/basic
```


### Managing Spinnaker

You can manage your Spinnaker installations with `kubectl`. 

#### Listing Spinnaker instances
```bash
$ kubectl get spinnakerservice --all-namespaces
NAMESPACE     NAME        VERSION
mynamespace   spinnaker   1.15.1
```

The short name `spinsvc` is also available.

#### Describing Spinnaker instances
```bash
$ kubectl -n mynamespace describe spinnakerservice spinnaker
```

#### Describing Spinnaker instances
Delete:
```bash
$ kubectl -n mynamespace deleted spinnakerservice spinnaker
spinnakerservice.spinnaker.io "spinnaker" deleted
```


## Spinnaker Configuration (TODO)

### `SpinnakerService`

The `SpinnakerService` points to the `configMap` with the configuration (see below).

### Spinnaker Configuration

## Architecture (TODO)
