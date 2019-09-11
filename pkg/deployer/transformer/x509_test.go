package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestTransformManifests_NewX509ServiceExposed(t *testing.T) {
	tr, spinSvc, _ := th.setupTransformer(&x509TransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = "service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"
	spinSvc.Spec.Expose.Service.Annotations = map[string]string{
		"service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http",
		"service.beta.kubernetes.io/aws-load-balancer-ssl-cert":         "arn::",
		"service.beta.kubernetes.io/aws-load-balancer-ssl-ports":        "80,443",
	}

	err := tr.TransformManifests(context.TODO(), nil, gen)
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

func TestTransformManifests_RemoveX509Service(t *testing.T) {
	x509Svc := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", x509Svc, t)
	x509Svc.Name = util.GateX509ServiceName
	fakeClient := fake.NewFakeClient(x509Svc)
	tr, spinSvc, hc := th.setupTransformerWithFakeClient(&x509TransformerGenerator{}, fakeClient, t)
	spinSvc.Spec.Expose.Type = "service"
	hc.Profiles = map[string]interface{}{}
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)
	x509Config, ok := gen.Config["gate-x509"]
	assert.True(t, ok)
	assert.Equal(t, x509Svc, x509Config.ToDelete[0])
}
