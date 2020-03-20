package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestTransformManifests_NewX509ServiceExposed(t *testing.T) {
	tr, spinSvc := th.setupTransformerFromSpinFile(&x509TransformerGenerator{}, "testdata/spinsvc_expose.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)
	spinSvc.GetExposeConfig().Service.PublicPort = 0

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", expected, t)
	expected.Name = "spin-gate-x509"
	expected.Spec.Ports[0].Name = "gate-x509"
	expected.Spec.Ports[0].Port = 443
	expected.Spec.Ports[0].TargetPort = intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: 8085,
	}
	assert.Equal(t, expected, gen.Config["gate-x509"].Service)
}

func TestTransformManifests_ExposedWithCustomPort(t *testing.T) {
	tr, spinSvc := th.setupTransformerFromSpinFile(&x509TransformerGenerator{}, "testdata/spinsvc_expose.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)
	spinSvc.GetExposeConfig().Service.PublicPort = 3333

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", expected, t)
	expected.Name = "spin-gate-x509"
	expected.Spec.Ports[0].Name = "gate-x509"
	expected.Spec.Ports[0].Port = 3333
	expected.Spec.Ports[0].TargetPort = intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: 8085,
	}
	assert.Equal(t, expected, gen.Config["gate-x509"].Service)
}

func TestTransformManifests_ExposedWithOverridenPort(t *testing.T) {
	s := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  spinnakerConfig:
    config: {}
    profiles:
      gate:
        default:
          apiPort: 8085
  expose:
    type: service
    service:
      type: LoadBalancer
      publicPort: 80
      annotations:
        "service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http"
      overrides: 
        gate-x509:
          publicPort: 5555
`
	tr, _ := th.setupTransformerFromSpinText(&x509TransformerGenerator{}, s, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", expected, t)
	expected.Name = "spin-gate-x509"
	expected.Spec.Ports[0].Name = "gate-x509"
	expected.Spec.Ports[0].Port = 5555
	expected.Spec.Ports[0].TargetPort = intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: 8085,
	}
	assert.Equal(t, expected, gen.Config["gate-x509"].Service)
}

func TestTransformManifests_RemoveX509Service(t *testing.T) {
	x509Svc := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", x509Svc, t)
	x509Svc.Name = util.GateX509ServiceName
	tr, spinSvc := th.setupTransformerFromSpinFile(&x509TransformerGenerator{}, "testdata/spinsvc_expose.yml", t, x509Svc)
	spinSvc.GetSpinnakerConfig().Profiles = map[string]interfaces.FreeForm{}
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)
	x509Config, ok := gen.Config["gate-x509"]
	assert.True(t, ok)
	assert.Equal(t, x509Svc, x509Config.ToDelete[0])
}
