package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/apps/v1beta2"
	"testing"
)

func TestTransformManifests_CustomServerPort(t *testing.T) {
	tr, _ := th.setupTransformer(&serverPortTransformerGenerator{}, "testdata/spinsvc_profile.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddDeploymentToGenConfig(gen, "gate", "testdata/input_deployment.yml", t)

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &v1beta2.Deployment{}
	test.ReadYamlFile("testdata/input_deployment.yml", expected, t)
	expected.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = int32(1111)
	expected.Spec.Template.Spec.Containers[0].ReadinessProbe.Exec.Command[4] = "http://localhost:1111/health"
	assert.Equal(t, expected, gen.Config["gate"].Deployment)
}
