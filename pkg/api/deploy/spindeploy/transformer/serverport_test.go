package transformer

import (
	"context"
	"testing"

	"github.com/armory/spinnaker-operator/pkg/api/generated"
	"github.com/armory/spinnaker-operator/pkg/api/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
)

func TestTransformManifests_CustomServerPort(t *testing.T) {
	tr, _ := th.SetupTransformerFromSpinFile(&ServerPortTransformerGenerator{}, "testdata/spinsvc_profile.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddDeploymentToGenConfig(gen, "gate", "testdata/input_deployment.yml", t)

	err := tr.TransformManifests(context.TODO(), gen)
	assert.Nil(t, err)

	expected := &v1.Deployment{}
	test.ReadYamlFile("testdata/input_deployment.yml", expected, t)
	expected.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = int32(1111)
	expected.Spec.Template.Spec.Containers[0].ReadinessProbe.Exec.Command[4] = "http://localhost:1111/health"
	assert.Equal(t, expected, gen.Config["gate"].Deployment)
}
