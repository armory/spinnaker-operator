# Spinnaker Operator for Kubernetes

We've announced the [Spinnaker Operator](https://blog.armory.io/spinnaker-operator/): a Kubernetes operator to deploy and manage Spinnaker with the tools you're used to. We're sharing configuration in this repository (code to come soon) to let the community evaluate it and provide feedback. 
Please let us know what would make your life easier when installing Spinnaker! You can use [GitHub issues](https://github.com/armory/spinnaker-operator/issues) for the time being.

*The operator is in alpha and its CRD may change quite a bit. It is actively being developed.*

## Requirements
The validating admission controller [requires](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites):
- Kubernetes server **v1.13+**
- Admission controllers enabled (`-enable-admission-plugins`)
- `ValidatingAdmissionWebhook` enabled in the kube-apiserver (should be the default)

Note: If you can't use the validation webhook, pass the `--disable-admission-controller` to the operator (like in `deploy/operator/basic/deployment.yaml`).

## Spinnaker installed in under a minute (or two)

For the impatient, more explanation can be found below.

```bash
# For a stable release (https://github.com/armory/spinnaker-operator/releases)
$ mkdir -p spinnaker-operator && cd spinnaker-operator
$ RELEASE=v0.2.0 bash -c 'curl -L https://github.com/armory/spinnaker-operator/releases/download/${RELEASE}/manifests.tgz | tar -xz'
 
# For the latest development work (master) 
$ git clone https://github.com/armory/spinnaker-operator.git && cd spinnaker-operator

# Install or update CRDs cluster wide
$ kubectl apply -f deploy/crds/

# Install operator in namespace spinnaker-operator, see below if you want a different namespace
$ kubectl create ns spinnaker-operator
$ kubectl -n spinnaker-operator apply -f deploy/operator/cluster

# Install Spinnaker in "spinnaker" namespace
$ kubectl create ns spinnaker
$ kubectl -n spinnaker apply -f deploy/spinnaker/basic

# Watch the install progress, check out the pods being created too!
$ kubectl -n spinnaker get spinsvc spinnaker -w
```

## What can you do with the Spinnaker Operator?

- Stop using Halyard commands: just `kubectl apply` your Spinnaker configuration. This includes support for local files.
- Expose Spinnaker to the outside world (via `LoadBalancer`). You can still disable that behavior if you prefer to manage ingresses and LBs yourself. 
- Deploy any version of Spinnaker. The operator is not tied to a particular version of Spinnaker. 
- Keep secrets separate from your config, store your config in `git`, and have an easy Gitops workflow.
- Validate your configuration before applying it (with webhook validation) 
- Store Spinnaker secrets in Kubernetes secrets
- Patch versions, accounts or any setting with `kustomize`. 
- Monitor the health of Spinnaker via `kubectl`
- Define Kubernetes accounts in `SpinnakerAccount` objects and store kubeconfig inline, in Kubernetes secrets, in s3, or gcs **[experimental]**

## Accounts CRD (experimental)
The Spinnaker Operator introduces a new CRD for Spinnaker accounts. A `SpinnakerAccount` is defined in an object - separate
from the main Spinnaker config - so its creation and maintenance can easily be automated.

The long term goal is to support all accounts (providers, CI, notifications, ...) but the first implementation deals with
Kubernetes accounts.

| Account type | Status |
|------------|----------|
| `Kubernetes` | alpha |

Read more at [Spinnaker accounts](doc/spinnaker-accounts.md)


## Operator Installation (detailed)
Download CRDs and example manifests from the [latest stable release](https://github.com/armory/spinnaker-operator/releases).
CRD and examples on `master` are unstable and subject to change but feedback is greatly appreciated.

**Breaking Change**: In 0.2.x+, the CRD no longer references a `configMap` but contains the whole configuration. 
It allows users to use `kustomize` to layer their Spinnaker changes and makes validation easier.    

### Step 1: Install CRDs

First we'll install the `SpinnakerService` and `SpinnakerAccount` CRDs:

```bash
$ mkdir -p spinnaker-operator && cd spinnaker-operator
$ tar -xvf operator-manifests.tgz .
$ kubectl apply -f deploy/crds/
```

Note: `SpinnakerAccount` CRD is optional.


### Step 2: Install Operator

There are two modes for the operator:
- **basic mode** to install Spinnaker in a single namespace without validating admission webhook
- **cluster mode** works across namespaces and requires a `ClusterRole` to perform validation

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
1. Decide and create the namespace the operator should live in. We suggest `spinnaker-operator`.
2. If you pick a different namespace than `spinnaker-operator`, edit the namespace in `deploy/operator/cluster/role_binding.yml`.
3. Run:

```bash
$ kubectl apply -n spinnaker-operator -f deploy/operator/cluster
```

## Spinnaker Installation

Once you've installed CRDs and operator, check out examples in `deploy/spinnaker/`. Below the 
`spinnaker-namespace` parameter refers to the namespace where you want to install
Spinnaker. It is likely different from  the operator's namespace.


### Example 1: Basic Install

**Important**: In `deploy/spinnaker/basic/spinnakerservice.yml`, change the `config.persistentStorage` section to point to an s3 bucket you own or use a different persistent storage.

`spinnakerservice.yml` currently points to version `1.17.1` but you can install any version of Spinnaker with the operator. Just change the version in the manifest.  

```bash
$ kubectl create ns <spinnaker-namespace>
$ kubectl -n <spinnaker-namespace> apply -f deploy/spinnaker/basic/spinnakerservice.yml
```

This configuration does not contain any connected accounts, just a persistent storage.

### Example 2: Install with all parameters

You'll find a more complete example under `deploy/spinnaker/basic/spinnakerservice.yml` with all parameters available.

### Example 3: Using Kustomize

Set your own values in `deploy/spinnaker/kustomize/kustomization.yml`, then:


```bash
$ kubectl create ns <spinnaker-namespace>
$ kustomize build deploy/spinnaker/kustomize/ | kubectl -n <spinnaker-namespace> apply -f -
```
 
## SpinnakerService options
See [all SpinnakerService options](doc/options.md).

## Uninstalling the operator
See [this section](doc/uninstalling.md).