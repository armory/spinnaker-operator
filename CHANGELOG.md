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

The `networking.k8s.io/v1` ingress is supported and queried.

For Gate, the operator also checks for the path and sets up Spinnaker to support relative path.

Another example with UI (Deck) at https://acme.com and API (Gate) at https://acme.com/api/v1
```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1
metadata:
  name: my-ingress
  namespace: spinnaker
spec:
  ingressClassName: nginx
  tls:
    - hosts: [ 'example.com', 'acme.com'] # That's how we know TLS is supported
  rules:
    - host: acme.com
      http:
        paths:
          - path: /api
            backend:
              service:
                name: spin-gate
                port:
                  name: http
          - path: /
            backend:
              service:
                name: spin-deck
                port:
                  number: 9000
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
