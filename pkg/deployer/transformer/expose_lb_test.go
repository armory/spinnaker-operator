package transformer

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestTransformManifests_ExposedNoOverrideUrl(t *testing.T) {
	tr, spinSvc, hc := th.setupTransformer(&exposeLbTransformerGenerator{}, t)
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
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedWithOverrideUrlChangingPort(t *testing.T) {
	tr, spinSvc, hc := th.setupTransformer(&exposeLbTransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = "service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"
	spinSvc.Spec.Expose.Service.Annotations = map[string]string{
		"service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http",
		"service.beta.kubernetes.io/aws-load-balancer-ssl-cert":         "arn::",
		"service.beta.kubernetes.io/aws-load-balancer-ssl-ports":        "80,443",
	}
	err := hc.SetHalConfigProp("security.apiSecurity.overrideBaseUrl", "https://my-api.spin.com")

	err = tr.TransformManifests(nil, hc, gen, nil)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", expected, t)
	expected.Spec.Ports[0].Port = int32(443)
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedAggregatedAnnotations(t *testing.T) {
	tr, spinSvc, hc := th.setupTransformer(&exposeLbTransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = "service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"
	spinSvc.Spec.Expose.Service.Annotations = map[string]string{
		"service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http",
		"service.beta.kubernetes.io/aws-load-balancer-ssl-cert":         "arn::",
	}
	spinSvc.Spec.Expose.Service.Overrides["gate"] = spinnakerv1alpha1.ExposeConfigServiceOverrides{
		Annotations: map[string]string{
			"service.beta.kubernetes.io/aws-load-balancer-ssl-ports": "80,443",
		},
	}

	err := tr.TransformManifests(nil, hc, gen, nil)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", expected, t)
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedServiceTypeOverridden(t *testing.T) {
	tr, spinSvc, hc := th.setupTransformer(&exposeLbTransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = "service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"
	spinSvc.Spec.Expose.Service.Overrides["gate"] = spinnakerv1alpha1.ExposeConfigServiceOverrides{
		Type: "NodePort",
	}

	err := tr.TransformManifests(nil, hc, gen, nil)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", expected, t)
	expected.Spec.Type = "NodePort"
	expected.Annotations = nil
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_NotExposed(t *testing.T) {
	tr, spinSvc, hc := th.setupTransformer(&exposeLbTransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = ""

	err := tr.TransformManifests(nil, hc, gen, nil)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", expected, t)
	expected.Annotations = nil
	expected.Spec.Type = "ClusterIP"
	assert.Equal(t, expected, gen.Config["gate"].Service)
}
