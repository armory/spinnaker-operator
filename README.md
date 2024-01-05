[![Coverage Status](https://coveralls.io/repos/github/armory/spinnaker-operator/badge.svg)](https://coveralls.io/github/armory/spinnaker-operator)

# Spinnaker Operator for Kubernetes

The [Spinnaker Operator](https://blog.armory.io/spinnaker-operator/) is a Kubernetes operator to deploy and manage Spinnaker using familiar tools. We're sharing it to let the community evaluate it and provide feedback.
Please let us know what would make your life easier when installing Spinnaker! You can use [GitHub issues](https://github.com/armory/spinnaker-operator/issues) for the time being.

## Benefits of Operator

- Stop using Halyard commands: just `kubectl apply` your Spinnaker configuration. This includes support for local files.
- Expose Spinnaker to the outside world (via `LoadBalancer`). You can still disable that behavior if you prefer to manage ingresses and LBs yourself. 
- Deploy any version of Spinnaker. Operator is not tied to a particular version of Spinnaker. 
- Keep secrets separate from your config. Store your config in `git` and have an easy Gitops workflow.
- Validate your configuration before applying it (by using kubernetes's webhook validation).
- Store Spinnaker secrets in Kubernetes secrets.
- Patch versions, accounts or any setting with `kustomize`.
- Monitor the health of Spinnaker through `kubectl`.
- Store kubeconfig inline, in [Kubernetes secrets](doc/managing-spinnaker.md#secrets-in-kubernetes-secrets), in S3, or GCS.
- Define Kubernetes accounts in `SpinnakerAccount` objects **[experimental]**
- Total control over generated manifests with [inline kustomization](doc/options.md#speckustomize)

## Requirements
The validating admission controller [requires](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites):
- Kubernetes server **v1.13+**
- Admission controllers enabled (`-enable-admission-plugins`)
- `ValidatingAdmissionWebhook` enabled in the kube-apiserver (should be the default)

Note: If you can't use the validation webhook, pass the `--disable-admission-controller` to the operator (like in `deploy/operator/basic/deployment.yaml`).

### Kubernetes' compatibility matrix

| Kubernetes / Spinnaker Operator | 1.3.x               | 1.2.x              |
|---------------------------------|---------------------|--------------------|
| 1.25                            | :white_check_mark:  | :x:                |
| 1.24                            | :white_check_mark:  | :x:                |
| 1.23                            | :white_check_mark:  | :x:                |
| 1.22                            | :white_check_mark:  | :x:                |
| 1.21                            | :white_check_mark:  | :white_check_mark: |

## Quick Start

This is a high-level view of the commands you need to run for those who want to jump right in. More explanation can be found in the sections after this one.

```bash
# Pick a release from https://github.com/armory/spinnaker-operator/releases (or clone the repo and use the master branch for the latest development work)
$ mkdir -p spinnaker-operator && cd spinnaker-operator
$ bash -c 'curl -L https://github.com/armory/spinnaker-operator/releases/latest/download/manifests.tgz | tar -xz'
 
# Install or update CRDs cluster wide
$ kubectl apply -f deploy/crds/

# Install operator in namespace spinnaker-operator, see below if you want a different namespace
$ kubectl create ns spinnaker-operator
$ kubectl -n spinnaker-operator apply -f deploy/operator/cluster

```

### Deploy Spinnaker

You can use our [Spinnaker kustomize patches](https://github.com/armory/spinnaker-kustomize-patches)
to deploy Spinnaker.

```shell
# Clone the project
git clone git@github.com:armory/spinnaker-kustomize-patches.git
cd spinnaker-kustomize-patches

# Delete default recipe
rm kustomization.yml

# Create symlink for oss recipe
ln -s ./recipes/kustomization-oss-minimum.yml kustomization.yml

# Create the spinnaker namespace
kubectl create ns spinnaker

# Build the kustomize template and deploy in kubernetes
kustomize build . | kubectl apply -f -

# Watch the install progress, check out the pods being created too!
$ kubectl -n spinnaker get spinsvc spinnaker -w
```

See [managing Spinnnaker](doc/managing-spinnaker.md)

## Accounts CRD (experimental)
The Spinnaker Operator introduces a new CRD for Spinnaker accounts. A `SpinnakerAccount` is defined in an object -- separate
from the main Spinnaker config -- so creation and maintenance can easily be automated.

The long term goal is to support all accounts (providers, CI, notifications, ...), but the first implementation deals with
Kubernetes accounts.

| Account type | Status |
|------------|----------|
| `Kubernetes` | alpha |

Read more at [Spinnaker accounts](doc/spinnaker-accounts.md).


## Operator Installation (detailed)
Download CRDs and example manifests from the [latest stable release](https://github.com/armory/spinnaker-operator/releases).
CRD and examples on `master` are unstable and subject to change, but feedback is greatly appreciated.

### Step 1: Install CRDs

First, we'll install the `SpinnakerService` and `SpinnakerAccount` CRDs:

```bash
$ mkdir -p spinnaker-operator && cd spinnaker-operator
$ tar -xvf manifests.tgz .
$ kubectl apply -f deploy/crds/
```

Note: `SpinnakerAccount` CRD is optional.


### Step 2: Install Operator

There are two modes for the operator:
- **Basic mode** installs Spinnaker into the same namespace as the Spinnaker Opertor without validating admission webhook.
- **Cluster mode** works across namespaces and requires a `ClusterRole` to perform validation.

The main difference between the two modes is that basic only requires a `Role` (vs a `ClusterRole`) and has no validating webhook.

Once installed, you should see a new deployment representing the Operator. The Operator watches for changes to the `SpinnakerService` objects. You can check on the status of the Operator using `kubectl`.

#### Basic install (no validating webhook)
Decide what namespace you want to use for Operator and create that namespace. We suggest `spinnaker-operator`.
To install Operator run:

```bash
$ kubectl apply -n spinnaker-operator -f deploy/operator/basic
```

#### Cluster install
To install Operator:
1. Decide what namespace you want to use for Operator and create that namespace. We suggest `spinnaker-operator`.
2. If you pick a different namespace than `spinnaker-operator`, edit the namespace in `deploy/operator/cluster/role_binding.yml`.
3. Run:

```bash
$ kubectl apply -n spinnaker-operator -f deploy/operator/cluster
```

If you use a namespace other than `spinnaker-operator`, replace `spinnaker-operator` with your namespace.

## Spinnaker Installation

Once you've installed CRDs and Operator, check out examples in `deploy/spinnaker/`. Below, the 
`spinnaker-namespace` parameter refers to the namespace where you want to install
Spinnaker. It is likely different from the operator's namespace.


### Example 1: Basic Install

In `deploy/spinnaker/basic/spinnakerservice.yml`, change the `config.persistentStorage` section to point to an S3 bucket you own or use a different persistent storage. Also make sure to update the Spinnaker version to the [desired version](https://www.spinnaker.io/community/releases/versions/#latest-stable). Since you installed Operator in `basic` mode, you must use the `spinnaker-operator` namespace. The permissions in `basic` mode are scoped to a single namespace so it doesn't see anything in other namespaces.

```bash
$ $ kubectl -n spinnaker-operator apply -f deploy/spinnaker/basic/spinnakerservice.yml
```

This configuration does not contain any connected accounts, just a persistent storage.

### Example 2: Install with all parameters

You'll find a more complete example under `deploy/spinnaker/complete/spinnakerservice.yml` with all parameters available.

### Example 3: Using Kustomize

Set your own values in `deploy/spinnaker/kustomize/kustomization.yml`, then:

```bash
$ kubectl create ns <spinnaker-namespace>
$ kustomize build deploy/spinnaker/kustomize/ | kubectl -n <spinnaker-namespace> apply -f -
```

## Install Spinnaker plugins

See the [Armory Operator Plugins reference](https://docs.armory.io/docs/installation/operator-reference/plugins/) and the [pf4jStagePlugin Deployment Using the Armory Operator](https://docs.armory.io/docs/plugin-guide/pf4j-deploy-example/) guide.

## Migrate from Halyard to Operator
See [how to migrate from Halyard to Spinnaker Operator](doc/migrate.md).
 
## SpinnakerService options
See [all SpinnakerService options](doc/options.md).

## Uninstalling the operator
See [this section](doc/uninstalling.md).
