# v1.4.0

## Features
- feat(k8s): Add support for GA ingress apiVersion in helm chart and operator (#287)

## Improvements
- chore(docker): Update alpine images operator + halyard (#292)
- chore(release): Update deployment manifest with specific release tag (#282)
- docs(k8s): Kubernetes compatibility matrix (#285)
- chore(release): Manifests update

# v1.3.1

## Improvements
- chore(docker): Update alpine images operator + halyard (#292) (#293)

## Dependencies
- Updated halyard version to operator-a6ac1d4

# v1.3.0

## Features
- feat(lambda/validation): Validation regarding the AWS Lambda using the GO SDK to get the Lambda Functions using the AWS provider credentials
- feat(cloudfoundry/validation): Add a CloudFoundry validation for each account (#222)
- feat(validator/aws): Add AWS account validator (#195)
- feat(health-check): Increase timeout and validate ready replicas for status (#219)
- feat(halyard): Bump version (#269)

## Bug Fixes
- fix(build): Kind unable to start control-plane (#279)
- fix(actions): Update metallb condition (#259)
- fix(it): Fix integration tests (#236)
- fix(expose): Override public service port (#210)
- fix(validation): Validate primary account for kubernetes provider (#209)
- fix(timeout): Avoid revalidation when patching Spinnaker Status (#192)
- fix(build): No push dev image to scan.connect.redhat.com (#184)
- Fix transforming k8s secrets (#262)

## Improvements
- chore(ci): Upgrading actions (#284)
- chore(ci): Skipping integration tests (#283)
- chore(dependency): Update upstream oss halyard version (#280, #278)
- chore(build): Split actions for PRs and master (#273)
- chore(build): Make sure forks can run tests, but not create releases (#268)
- chore(halyard): Updated halyard version (#232, #213, #200, #197)
- chore(cve): Fix for CVE-2020-13757 (#193)
- update(kind): Update yml reference files (#258, #257)
- update(operator): Update files for the new API of k8s v1.22 (#252) (#256)
- doc(fix): Add Plugins section to README (#220)

## Dependencies
- Updated halyard version to operator-b135799
- Support for Kubernetes v1.22 API changes

# v1.2.5

## Improvements
- chore(release): v1.2.5 (#233)

## Dependencies
- Updated halyard version to operator-ccae06e

# v1.2.4

## Features
- feat(health-check): Increase timeout and validate ready replicas for status (#219) (#221)

# v1.2.3

## Improvements
- chore(release): v1.2.3 (#217)

# v1.2.2

## Improvements
- chore(halyard): Updated halyard version (#200) (#201)

## Dependencies
- Updated halyard version to operator-7162184

# v1.2.1

## Improvements
- chore(release): v1.2.1 (#198)

## Dependencies
- Updated halyard version to operator-8e0406f

# v1.2.0

## Features
- feat(health-check): Check spinnaker status (#168)
- feat(ubi): Add build for UBI images (#158)

## Bug Fixes
- fix(build): No push dev image to scan.connect.redhat.com (#184) (#185)
- fix(ingress): Fix panic when overriding endpoint with ingress (#181)
- fix(test): Fix integration tests (#179, #171)
- fix(k8s-context): Use the context passed in SpinnakerService when validating Kubernetes accounts (#173)
- fix(ingress): Support ingress with load balancer IP (GCE/bare metal) (#170)
- fix(ubi): Fix UBI run issue (#169)
- fix(test): Added ingress permissions to role used in tests (#155)
- fix(expose/ingress): Solve issue for spinsvc status URL (#154)

## Improvements
- chore(halyard): Update version (#183)
- chore(release): Updated halyard version (#174)
- chore(license): Update Copyright section (#162)
- chore(coverage): Add code coverage (#161)
- chore(mergify): Add mergify config (#157)
- chore(release): Update changelog

## Dependencies
- Updated halyard version to operator-c1d641c

# Unreleased (1.1.2)

- chore: Update halyard version.
- fix: Validation Kubernetes accounts using the context passed on Spinnaker Service.
- refactor: Introducing a better way to check spinnaker health validating correct status of each pod.

# v1.1.0

Breaking change:
- `roles.yaml` has changed for `Ingress` support. You only need to update if you want to use `Ingress`. 

## Ingress Support
`spec.expose.type: ingress`. When `ingress` is selected, the operator will try find an ingress rule 
in the same namespace as Spinnaker that point to Gate or Deck. It will then compute these services' hostnames
using (`spec.rules[].host` or `status.loadBalancer.ingress[0].hostname`).

Both `extensions` and `networking.k8s.io` ingresses are supported and queried.

For Gate, the operator also checks for the path and sets up Spinnaker to support relative path.

e.g. the following will setup Spinnaker's UI (Deck) at http://acme.com and API (Gate) at http://acme.com/api
```yaml
kind: Ingress
apiVersion: extensions/v1beta1
metadata:
  name: my-ingress
  namespace: spinnaker
spec:
  rules:
    - http:
        paths:
          - path: /api
            backend:
              serviceName: spin-gate
              servicePort: http
          - path: /
            backend:
              serviceName: spin-deck
              servicePort: 9000
status:
  loadBalancer:
    ingress:
      - hostname: acme.com
```

Another example with UI (Deck) at https://acme.com and API (Gate) at https://acme.com/api/v1
```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: my-ingress
  namespace: spinnaker
spec:
  tls:
    - hosts: [ 'example.com', 'acme.com'] # That's how we know TLS is supported
  rules:
    - host: acme.com
      http:
        paths:
          - path: /api
            backend:
              serviceName: spin-gate
              servicePort: http
          - path: /
            backend:
              serviceName: spin-deck
              servicePort: 9000
```
 
Note: Roles have changed to allow for ingress list.

# Others
- refactor: transformer and (change) detectors will now be organized by functionality and kept in different modules.
- fix: a long time bug in `FreeForm` is also fixed. It was causing transformers that attempts to modify the config (or profiles) in memory to also leak the change into the operator's informer cache.
- fix: Validation webhook now patches the status. We cannot return the patches directly because we're changing the status. That should fix some validation errors trying to apply a new `SpinnakerService`
- fix: Validation service has ports named for Istio support
- fix: Crash when using `SpinnakerAccount` with sharded services ("HA mode")

# v1.0.0
TODO: need to backfill this
