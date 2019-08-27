package transformer

import (
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestTransformManifests_NewX509ServiceExposed(t *testing.T) {
	tr, spinSvc, hc := th.setupTransformer(&x509TransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = "service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"
	spinSvc.Spec.Expose.Service.Annotations = map[string]string{
		"service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http",
		"service.beta.kubernetes.io/aws-load-balancer-ssl-cert":         "arn::",
		"service.beta.kubernetes.io/aws-load-balancer-ssl-ports":        "80,443",
	}

	err := tr.TransformManifests(nil, hc, gen, nil)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", expected, t)
	expected.Name = "spin-gate-x509"
	expected.Spec.Ports[0].Name = "gate-x509"
	expected.Spec.Ports[0].Port = 8085
	expected.Spec.Ports[0].TargetPort = intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: 8085,
	}
	assert.Equal(t, expected, gen.Config["gate-x509"].Service)
}

func TestTransformManifests_NewX509ServiceNotExposed(t *testing.T) {
	tr, _, hc := th.setupTransformer(&x509TransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)

	err := tr.TransformManifests(nil, hc, gen, nil)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("input_service.json", expected, t)
	expected.Name = "spin-gate-x509"
	expected.Spec.Ports[0].Name = "gate-x509"
	expected.Spec.Ports[0].Port = 8085
	expected.Spec.Ports[0].TargetPort = intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: 8085,
	}
	assert.Equal(t, expected, gen.Config["gate-x509"].Service)
}
