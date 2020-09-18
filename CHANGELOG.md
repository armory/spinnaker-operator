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
