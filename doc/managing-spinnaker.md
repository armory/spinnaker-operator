# Managing Spinnaker

### Installing Spinnaker

```bash
kubectl -n spinnaker apply -f spinnakerservice.yml 
```

You can manage your Spinnaker installations with `kubectl`. [Detailed information about the SpinnakerService CRD fields](./options.md)

### Listing Spinnaker instances
```bash
$ kubectl get spinsvc --all-namespaces
NAMESPACE   NAME        VERSION   LASTCONFIGURED   STATUS   SERVICES   URL
spinnaker   spinnaker   1.16.2    114m             OK       8          http://myloadbalancer.us-west-2.elb.amazonaws.com
```

### Describing Spinnaker instances
```bash
$ kubectl -n mynamespace describe spinnakerservice spinnaker
Name:         spinnaker
Namespace:    spinnaker
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"spinnaker.io/v1alpha2","kind":"SpinnakerService","metadata":{"annotations":{},"name":"spinnaker","namespace":"spinnaker"},"...
API Version:  spinnaker.io/v1alpha2
Kind:         SpinnakerService
Metadata:
  Creation Timestamp:  2019-11-01T16:21:09Z
  Generation:          27
  Resource Version:    13732856
  Self Link:           /apis/spinnaker.io/v1alpha2/namespaces/spinnaker/spinnakerservices/spinnaker
  UID:                 9cf793f3-fcc3-11e9-8adb-0a33131e8c2c
Spec:
  Accounts:
    Enabled:  true
  Expose:
    Service:
      Annotations:
        service.beta.kubernetes.io/aws-load-balancer-backend-protocol:  http
      Overrides:
      Type:  LoadBalancer
    Type:    service
    ...
Status:
  API URL:  http://myapiloadbalancer.us-west-2.elb.amazonaws.com
  Last Deployed:
    Config:
      Hash:             e3678e9c003c10f8ecb1098802dbfda1
      Last Updated At:  2019-11-06T01:49:53Z
    account-Kubernetes-myaccount:
      Hash:             37a6259cc0c1dae299a7866489dff0bd
      Last Updated At:  2019-11-06T03:01:04Z
  Service Count:        8
  Services:
    ...
  Status:            OK
  Ui URL:            http://myloadbalancer.elb.amazonaws.com
  Version:           1.16.2
Events:              <none>
```

### Deleting Spinnaker instances
Delete:

```bash
$ kubectl -n mynamespace delete spinnakerservice spinnaker
spinnakerservice.spinnaker.io "spinnaker" deleted
```


# Secrets
When it comes to storing secrets, you have several options each with their own pros and cons. Generally speaking, 
you should pick the method that matches your workflow the best. There's no significant performance differences between each option:

## Secrets in a cloud provider storage (s3, s3-like, gcs)
This method lets you store secrets externally. You can then manage access as any other bucket you manage.

Please refer to [the Spinnaker's documentation](https://www.spinnaker.io/reference/halyard/secrets/) for more details.

## Secrets in Kubernetes secrets
This method is only available via the Operator at this time. It is similar to the one above with a different syntax:

`encrypted:k8s!n:<secret name>!k:<key under which the secret is stored>`

Note that for security reason, Spinnaker can only access secrets stored in its own namespace (which could be different
than the operator's).

Example:

```yaml
spec:
  spinnakerConfig:
    config:
      persistentStorage:
        s3:
          accessKeyId: <my access key>
          secretAccessKey: encrypted:k8s!n:spinnaker-secrets!k:aws-access-key
      providers:
        kubernetes:
          accounts:
          - name: myaccount
            kubeconfigFile: encrypted:k8s!n:spinnaker-secrets!k:myaccount-kubeconfig
            ... 
``` 


