package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"testing"
)

func TestStrategicMerge(t *testing.T) {
	cases := []struct {
		name     string
		spinsvc  string
		expected func(*testing.T, *appsv1.Deployment)
	}{
		{
			"one strategic merge as json",
			`
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  kustomize:
    gate:
      deployment:
        patchesStrategicMerge:
          - |
            { "spec": { "selector": { "matchLabels": { "newlabel": "blah" }}}}
`,
			func(t *testing.T, dep *appsv1.Deployment) {
				assert.Equal(t, "blah", dep.Spec.Selector.MatchLabels["newlabel"])
			},
		},
		{
			"one strategic merge as yaml",
			`
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  kustomize:
    gate:
      deployment:
        patchesStrategicMerge:
          - |
            spec:
             selector:
               matchLabels:
                 newlabel: blah 
`,
			func(t *testing.T, dep *appsv1.Deployment) {
				assert.Equal(t, "blah", dep.Spec.Selector.MatchLabels["newlabel"])
			},
		},
		{
			"one json merge as yaml",
			`
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  kustomize:
    gate:
      deployment:
        patches:
          - |
            spec:
              template:
                spec:
                  containers:
                  - name: other
`,
			func(t *testing.T, dep *appsv1.Deployment) {
				if assert.Equal(t, 1, len(dep.Spec.Template.Spec.Containers)) {
					assert.Equal(t, "other", dep.Spec.Template.Spec.Containers[0].Name)
				}
			},
		},
		{
			"one strategic merge as yaml - add a container",
			`
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  kustomize:
    gate:
      deployment:
        patchesStrategicMerge:
          - |
            spec:
              template:
                spec:
                  containers:
                  - name: other
`,
			func(t *testing.T, dep *appsv1.Deployment) {
				assert.Equal(t, 2, len(dep.Spec.Template.Spec.Containers))
			},
		},
		{
			"json6902 patch as yaml",
			`
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  kustomize:
    gate:
      deployment:
        patchesJson6902: |
          - op: replace
            path: /spec/template/spec/containers/0/image
            value: nginx 
`,
			func(t *testing.T, dep *appsv1.Deployment) {
				if assert.Equal(t, 1, len(dep.Spec.Template.Spec.Containers)) {
					assert.Equal(t, "nginx", dep.Spec.Template.Spec.Containers[0].Image)
				}
			},
		},
		{
			"json6902 patch as yaml, remove volume mount",
			`
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  kustomize:
    gate:
      deployment:
        patchesJson6902: |
          - op: remove
            path: /spec/template/spec/containers/0/volumeMounts/1
`,
			func(t *testing.T, dep *appsv1.Deployment) {
				if assert.Equal(t, 1, len(dep.Spec.Template.Spec.Containers)) {
					assert.Equal(t, 1, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts))
				}
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, _ := th.SetupTransformerFromSpinText(&PatchTransformerGenerator{}, c.spinsvc, t)
			gen := &generated.SpinnakerGeneratedConfig{}
			test.AddDeploymentToGenConfig(gen, "gate", "testdata/input_deployment.yml", t)
			err := p.TransformManifests(context.TODO(), gen)
			if assert.Nil(t, err) {
				c.expected(t, gen.Config["gate"].Deployment)
			}
		})
	}
}
