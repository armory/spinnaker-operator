# Uninstalling the Operator

If for some reason the operator needs to be uninstalled/deleted, there are still two ways in which Spinnaker itself can be prevented from being deleted, explained in the following sections.

### Replacing the operator with Halyard

First you need to export Spinnaker configuration settings to a format that Halyard understands: 
1. From the `SpinnakerService` manifest, copy the contents of `spec.spinnakerConfig.config` to its own file named `config`, and save it with the following structure:
```
currentDeployment: default
deploymentConfigurations:
- name: default
  <<CONTENT HERE>> 
```
2. For each entry in `spec.spinnakerConfig.profiles`, copy it to its own file inside a `profiles` folder with a `<entry-name>-local.yml` name.
3. For each entry in `spec.spinnakerConfig.service-settings`, copy it to its own file inside a `service-settings` folder with a `<entry-name>.yml` name.
4. For each entry in `spec.spinnakerConfig.files`, copy it to its own file inside a directory structure following the name of the entry with double underscores (__) replaced by a path separator. Example: an entry named `profiles__rosco__packer__example-packer-config.json` would produce the file `profiles/rosco/packer/example-packer-config.json`.

At the end, you would have the following directory tree:
```
config
default/
  profiles/
  service-settings/
```

After that, you can put these files in your Halyard home directory and deploy Spinnaker running `hal deploy apply`.

Finally you can delete the operator and their CRDs from the Kubernetes cluster.

```bash
$ kubectl delete -n <namespace> -f deploy/operator/<installation type>
$ kubectl delete -f deploy/crds/
```

### Removing operator ownership from Spinnaker resources

You can execute the following script to remove ownership of Spinnaker resources, where `NAMESPACE` is the namespace where Spinnaker is installed:
```bash
NAMESPACE=
for rtype in deployment service
do
    for r in $(kubectl -n $NAMESPACE get $rtype --selector=app=spin -o jsonpath='{.items[*].metadata.name}') 
    do
        kubectl -n $NAMESPACE patch $rtype $r --type json -p='[{"op": "remove", "path": "/metadata/ownerReferences"}]'
    done
done
```
Finally you can delete the operator and their CRDs from the Kubernetes cluster.
```bash
$ kubectl delete -n <namespace> -f deploy/operator/<installation type>
$ kubectl delete -f deploy/crds/
```
