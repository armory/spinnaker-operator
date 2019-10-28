# Spinnaker Operator for Kubernetes

We've announced the [Spinnaker Operator](https://blog.armory.io/spinnaker-operator/): a Kubernetes operator to deploy and manage Spinnaker with the tools you're used to. We're sharing configuration in this repository (code to come soon) to let the community evaluate it and provide feedback. 
Please let us know what would make your life easier when installing Spinnaker! You can use [GitHub issues](https://github.com/armory/spinnaker-operator/issues) for the time being.

*The operator is in alpha and its CRD may change quite a bit. It is actively being developed.*

## Requirements
The validating admission controller [requires](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites):
- Kubernetes server **v1.13+**
- Admission controllers enabled (`-enable-admission-plugins`)
- `ValidatingAdmissionWebhook` enabled in the kube-apiserver (should be the default)

Note: If you can't use the validation webhook, pass the `--without-admission-controller` to the operator (like in `deploy/operator/basic/deployment.yaml`).

## Installation
Download CRDs and example manifests from the [latest stable release](https://github.com/armory/spinnaker-operator/releases).
CRD and examples on `master` are unstable and subject to change.

**Breaking Change**: In 0.2.x+, the CRD no longer references a `configMap` but contains the whole configuration. 
It allows users to use `kustomize` to layer their Spinnaker changes and makes validation easier.    

### Operator Installation

First we'll install the `SpinnakerService` CRD:

```bash
$ mkdir -p spinnaker-operator && cd spinnaker-operator
$ tar -xvf operator-manifests.tgz .
$ kubectl apply -f deploy/crds/spinnaker_v1alpha1_spinnakerservice_crd.yaml
```

There are two modes for the operator:
- basic mode to install Spinnaker in a single namespace without validating admission webhook
- cluster mode works across namespaces and requires a `ClusterRole` to perform validation

The main difference between the two modes is that basic only requires a `Role` (vs a `ClusterRole`) and has no validating webhook.

Once installed you should see a new deployment representing the operator. The operator watches for changes to the `SpinnakerService` objects. You can check on the status of the operator using `kubectl`.

#### Basic install (no validating webhook)
To install the operator run:

```bash
$ kubectl apply -n <namespace> -f deploy/operator/basic
```

`namespace` is the namespace where you want the operator to live and deploy to.

#### Cluster install
To install the operator:
1. Edit the namespace in `deploy/operator/cluster/role_binding.yml` to be the namespace where you want the operator to live.
2. Run:

```bash
$ kubectl apply -n <namespace> -f deploy/operator/cluster
```

### Spinnaker Installation

Once you've installed the operator, you can install Spinnaker by making a configuration (`configMap`). Check out examples in `deploy/spinnaker/SpinnakerService.yml`. If you prefer to use `kustomize`, we've added some kustomization in `deploy/spinnaker/` (WIP)


#### Example 1: Installing version 1.15.1

**Important**: In `deploy/spinnaker/SpinnakerService.yml`, change the `config.persistentStorage` section to point to an s3 bucket you own or use a different persistent storage.


```bash
$ kubectl -n <namespace> apply -f deploy/spinnaker/SpinnakerService.yml
```

This configuration does not contain any connected accounts, just a persistent storage.

#### Example 2: Using Kustomize (TODO)

Set your own values in `deploy/spinnaker/kustomization.yml`, then:


```bash
$ kustomize build deploy/spinnaker/ | kubectl -n <namespace> apply -f -
```

Or if using `kubectl` version 1.14+:
```bash
$ kubectl -n <namespace> apply -f deploy/spinnaker/
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

#### Deleting Spinnaker instances
Delete:
```bash
$ kubectl -n mynamespace deleted spinnakerservice spinnaker
spinnakerservice.spinnaker.io "spinnaker" deleted
```


## Configuring Spinnaker

Detailed information about the SpinnakerService CRD fields and how to configure Spinnaker can be found [in the wiki](https://github.com/armory/spinnaker-operator/wiki/SpinnakerService-CRD)


