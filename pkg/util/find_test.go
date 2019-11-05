package util

import (
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	"testing"
)

func TestGetMountedSecretNameInDeployment(t *testing.T) {
	s := `
apiVersion: apps/v1
kind: Deployment
spec:
  selector: null
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers:
      - name: monitoring
        resources: {}
        volumeMounts:
        - mountPath: /opt/spinnaker/config
          name: test1
        - mountPath: /opt/monitoring
          name: test2
      - name: clouddriver
        resources: {}
        volumeMounts:
        - mountPath: /opt/spinnaker/config
          name: test3
        - mountPath: /opt/monitoring
          name: test1
      volumes:
      - name: test1
        secret:
          secretName: val1
      - name: test2
        secret:
          secretName: val2
      - name: test3
        secret:
          secretName: val3
status: {}`

	d := &v1.Deployment{}
	if assert.Nil(t, yaml.Unmarshal([]byte(s), d)) {
		v := GetMountedSecretNameInDeployment(d, "clouddriver", "/opt/spinnaker/config")
		assert.Equal(t, "val3", v)
	}
}
