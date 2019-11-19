# SpinnakerService Options

You can find a skeleton of `SpinnakerService` at [deploy/spinnaker/complete/spinnakerservice.yml](../deploy/spinnaker/complete/spinnakerservice.yml) 

```yaml
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  # spec.spinnakerConfig - This section is how to specify configuration spinnaker
  spinnakerConfig:
    # spec.spinnakerConfig.config - This section contains the contents of a deployment found in a halconfig .deploymentConfigurations[0]
    config:
      version: 1.17.1   # the version of Spinnaker to be deployed
      persistentStorage:
        persistentStoreType: s3
        s3:
          bucket: mybucket # change me
          rootFolder: front50

    # spec.spinnakerConfig.profiles - This section contains the YAML of each service's profile
    profiles:
      clouddriver: {} # is the contents of ~/.hal/default/profiles/clouddriver.yml
      # deck has a special key "settings-local.js" for the contents of settings-local.js
      deck:
        # settings-local.js - contents of ~/.hal/default/profiles/settings-local.js
        # Use the | YAML symbol to indicate a block-style multiline string
        settings-local.js: |
          window.spinnakerSettings.feature.kustomizeEnabled = true;
          window.spinnakerSettings.feature.artifactsRewrite = true;
      echo: {}    # is the contents of ~/.hal/default/profiles/echo.yml
      fiat: {}    # is the contents of ~/.hal/default/profiles/fiat.yml
      front50: {} # is the contents of ~/.hal/default/profiles/front50.yml
      gate: {}    # is the contents of ~/.hal/default/profiles/gate.yml
      igor: {}    # is the contents of ~/.hal/default/profiles/igor.yml
      kayenta: {} # is the contents of ~/.hal/default/profiles/kayenta.yml
      orca: {}    # is the contents of ~/.hal/default/profiles/orca.yml
      rosco: {}   # is the contents of ~/.hal/default/profiles/rosco.yml

    # spec.spinnakerConfig.service-settings - This section contains the YAML of the service's service-setting
    # see https://www.spinnaker.io/reference/halyard/custom/#tweakable-service-settings for available settings
    service-settings:
      clouddriver: {}
      deck: {}
      echo: {}
      fiat: {}
      front50: {}
      gate: {}
      igor: {}
      kayenta: {}
      orca: {}
      rosco: {}

    # spec.spinnakerConfig.files - This section allows you to include any other raw string files not handle above.
    # The KEY is the filepath and filename of where it should be placed
    #   - Files here will be placed into ~/.hal/default/ on halyard
    #   - __ is used in place of / for the path separator
    # The VALUE is the contents of the file.
    #   - Use the | YAML symbol to indicate a block-style multiline string
    #   - We currently only support string files
    #   - NOTE: Kubernetes has a manifest size limitation of 1MB
    files:
  #      profiles__rosco__packer__example-packer-config.json: |
  #        {
  #          "packerSetting": "someValue"
  #        }
  #      profiles__rosco__packer__my_custom_script.sh: |
  #        #!/bin/bash -e
  #        echo "hello world!"


  # spec.expose - This section defines how Spinnaker should be publicly exposed
  expose:
    type: service  # Kubernetes LoadBalancer type (service/ingress), note: only "service" is supported for now
    service:
      type: LoadBalancer

      # annotations to be set on Kubernetes LoadBalancer type
      # they will only apply to spin-gate, spin-gate-x509, or spin-deck
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
        # uncomment the line below to provide an AWS SSL certificate to terminate SSL at the LoadBalancer
        #service.beta.kubernetes.io/aws-load-balancer-ssl-cert: arn:aws:acm:us-west-2:9999999:certificate/abc-123-abc

      # provide an override to the exposing KubernetesService
      overrides:
      # Provided below is the example config for the Gate-X509 configuration
#        deck:
#          annotations:
#            service.beta.kubernetes.io/aws-load-balancer-ssl-cert: arn:aws:acm:us-west-2:9999999:certificate/abc-123-abc
#            service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
#        gate:
#          annotations:
#            service.beta.kubernetes.io/aws-load-balancer-ssl-cert: arn:aws:acm:us-west-2:9999999:certificate/abc-123-abc
#            service.beta.kubernetes.io/aws-load-balancer-backend-protocol: https  # X509 requires https from LoadBalancer -> Gate
#       gate-x509:
#         annotations:
#           service.beta.kubernetes.io/aws-load-balancer-backend-protocol: tcp
#           service.beta.kubernetes.io/aws-load-balancer-ssl-cert: null
#         publicPort: 443
```

## `metadata.name`
This is the name of your Spinnaker service. You'll use that name to view, edit, or delete Spinnaker.

Example with a `prod` name:
```bash
$ kubectl get spinsvc prod
```

Note: We use `spinsvc` for brevity. You can also use `spinnakerservices.spinnaker.io`. 

## `.spec.spinnakerConfig.config`

It supported the deployment content found in a halconfig `.deploymentConfigurations[0]`.

For instance, given the following:

.`~/.hal/config` file
```yaml
currentDeployment: default
deploymentConfigurations:
- name: default
  version: 1.17.1
  persistentStorage:
    persistentStoreType: s3
    s3:
      bucket: mybucket
      rootFolder: front50
```

We'd get the following `spec.spinnakerConfig`:

```yaml
spec:
  spinnakerConfig:
    config:
      version: 1.15.1
      persistentStorage:
        persistentStoreType: s3
        s3:
          bucket: mybucket
          rootFolder: front50
```    

## `.spec.spinnakerConfig.profiles`

This section contains each service profile. This is the equivalent of a `~/.hal/default/profiles/<service>-local.yml`

For example:
```yaml
spec:
  spinnakerConfig:
    config:
    ...
    profiles:
      gate:
        default:
          apiPort: 8085
```

NOTE: Deck's profile is a string under the key `settings-local.js`:
```yaml
spec:
  spinnakerConfig:
    config:
    ...
    profiles:
      deck:
        settings-local.js: |
          window.spinnakerSettings.feature.artifactsRewrite = true;
```

## `spec.expose`
This section contains configuration for exposing Spinnaker. It is optional, if you omit it
no load balancer will be created (or deleted if you remove it). 


### `spec.expose.type`
This defines how How Spinnaker will be exposed. Only `service` is currently supported for using Kubernetes services.

#### `spec.expose.service`
Service Configuration

##### `spec.expose.service.type`
Matches a valid kubernetes service type (i.e. `LoadBalancer`, `NodePort`, `ClusterIP`).

IMPORTANT: `LoadBalancer` is the only supported type currently.

##### `spec.expose.service.annotations`
Map containing any annotation to be added to Gate (API) and Deck (UI) services.

##### `spec.expose.service.overrides`
Map with key: Spinnaker service name (`gate` or `deck`), and value: structure for overriding the service type and specifying extra annotations.
By default, all services will receive the same annotations.
You can override annotations for a Deck (UI) or Gate (API) services.

## `spec.validation`

Contains validation options that apply to all validations performed by the operator:

### `spec.validation.failOnError`
Defaults to true. If false, the validation is run and its result logged but the service is always valid

### `spec.validation.failFast`
Defaults to false, if true, validation will stop at the first error

### `spec.validation.frequencySeconds`
Optional parameter. Defines a grace period before a validation is re-run.

For instance, if you define a value of `120` and edit the `SpinnakerService` without changing an account within 120 seconds,
the validation on that account won't be run again.

Note: In the future, this will run validation while Spinnaker is running.

### `spec.validation.providers`, `spec.validation.ci`, `spec.validation.metricStores`, `spec.validation.persistentStorage`, `spec.validation.notifications`
Optional maps of validation settings specific to certain providers/CI/etc. Supported settings are:
- `enabled`: to turn off validation
- `failOnError`
- `frequencySeconds` 

Example to disable all Kubernetes account validation:
```yaml
spec:
  validation:
    providers:
      kubernetes:
        enabled: false
``` 

## `spec.accounts` 
Support for `SpinnakerAccount` CRD

### `spec.accounts.enabled`
Defaults ot `false`. If `true`, the `SpinnakerService` will use all `SpinnakerAccount` objects enabled.

Note: For now, accounts can be defined either in the config or as `SpinnakerAccount`. For instance,
if you add a `SpinnakerAccount` of type `kubernetes`, accounts you've defined elsewhere won't be loaded.

### `spec.accounts.dynamic` (experimental)
Defaults to `false`. If `true`, `SpinnakerAccount` objects will be made available to 
Spinnaker as the account is applied - without redeploying any service.


# Expose Examples

## Exposing Spinnaker with LoadBalancer services 

```yaml
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  expose:
    type: service
    service:
      type: LoadBalancer
      annotations:
        "service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http"
        "service.beta.kubernetes.io/aws-load-balancer-ssl-ports": "80,443"
        "service.beta.kubernetes.io/aws-load-balancer-ssl-cert": "arn:aws:acm:us-west-2:xxxxxxxxxxxx:certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

Above manifest file will generate these two services:

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: 80,443
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert": arn:aws:acm:us-west-2:xxxxxxxxxxxx:certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  labels:
    app: spin
    cluster: spin-deck
  name: spin-deck
spec:
  ports:
 - name: deck-tcp
   nodePort: xxxxx
   port: 9000
   protocol: TCP
   targetPort: 9000
  selector:
   app: spin
   cluster: spin-deck
  sessionAffinity: None
  type: LoadBalancer
```


```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: 80,443
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert": arn:aws:acm:us-west-2:xxxxxxxxxxxx:certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  labels:
     app: spin
     cluster: spin-gate
  name: spin-gate
spec:
  ports:
  - name: gate-tcp
    nodePort: xxxxx
    port: 8084
    protocol: TCP
    targetPort: 8084
  selector:
    app: spin
    cluster: spin-gate
  sessionAffinity: None
  type: LoadBalancer
```

## Exposing Spinnaker, different service types for Deck (UI) and Gate (API)

```yaml
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  expose:
    type: service
    service:
      type: LoadBalancer
      annotations:
        "service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http"
        "service.beta.kubernetes.io/aws-load-balancer-ssl-ports": "80,443"
        "service.beta.kubernetes.io/aws-load-balancer-ssl-cert": "arn:aws:acm:us-west-2:xxxxxxxxxxxx:certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
      overrides:
        gate:
          type: NodePort
```

Above manifest file will generate these two services:

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: 80,443
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert": arn:aws:acm:us-west-2:xxxxxxxxxxxx:certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  labels:
    app: spin
    cluster: spin-deck
  name: spin-deck
  spec:
  ports:
  - name: deck-tcp
    nodePort: xxxxx
    port: 9000
    protocol: TCP
    targetPort: 9000
  selector:
    app: spin
    cluster: spin-deck
  sessionAffinity: None
  type: LoadBalancer
```

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: 80,443
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert": arn:aws:acm:us-west-2:xxxxxxxxxxxx:certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  labels:
    app: spin
    cluster: spin-gate
  name: spin-gate
spec:
  ports:
  - name: gate-tcp
    nodePort: xxxxx
    port: 8084
    protocol: TCP
    targetPort: 8084
  selector:
    app: spin
    cluster: spin-gate
  sessionAffinity: None
  type: NodePort
```

## Exposing Spinnaker, different annotations for Deck (UI) and Gate (API)

```yaml
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  expose:
    type: service
    service:
      type: LoadBalancer
      annotations:
        "service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http"
        "service.beta.kubernetes.io/aws-load-balancer-ssl-ports": "80,443"
        "service.beta.kubernetes.io/aws-load-balancer-ssl-cert": "arn:aws:acm:us-west-2:xxxxxxxxxxxx:certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
      overrides:
        gate:
          annotations:
            "service.beta.kubernetes.io/aws-load-balancer-internal": "true"
```

Above manifest file will generate these two services:

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: 80,443
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert": arn:aws:acm:us-west-2:xxxxxxxxxxxx:certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  labels:
    app: spin
    cluster: spin-deck
  name: spin-deck
spec:
  ports:
  - name: deck-tcp
    nodePort: xxxxx
     port: 9000
     protocol: TCP
     targetPort: 9000
  selector:
     app: spin
     cluster: spin-deck
  sessionAffinity: None
  type: LoadBalancer
```

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: 80,443
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert": arn:aws:acm:us-west-2:xxxxxxxxxxxx:certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    service.beta.kubernetes.io/aws-load-balancer-internal: true
  labels:
    app: spin
    cluster: spin-gate
  name: spin-gate
spec:
  ports:
 - name: gate-tcp
    nodePort: xxxxx
    port: 8084
    protocol: TCP
    targetPort: 8084
  selector:
    app: spin
    cluster: spin-gate
  sessionAffinity: None
  type: Loadbalancer
```


## Exposing Spinnaker with X509
```yaml
spec:
  config:
    profiles:
      gate:
        default:
          apiPort: 8085  
  expose:
    type: service
    service:
      type: LoadBalancer

      annotations:
        service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http

      overrides:
      # Provided below is the example config for the Gate-X509 configuration
        deck:
          annotations:
            service.beta.kubernetes.io/aws-load-balancer-ssl-cert: arn:aws:acm:us-west-2:9999999:certificate/abc-123-abc
            service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
        gate:
          annotations:
            service.beta.kubernetes.io/aws-load-balancer-ssl-cert: arn:aws:acm:us-west-2:9999999:certificate/abc-123-abc
            service.beta.kubernetes.io/aws-load-balancer-backend-protocol: https  # X509 requires https from LoadBalancer -> Gate
       gate-x509:
         annotations:
           service.beta.kubernetes.io/aws-load-balancer-backend-protocol: tcp
           service.beta.kubernetes.io/aws-load-balancer-ssl-cert: null
         publicPort: 443
```
