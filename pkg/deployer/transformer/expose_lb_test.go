package transformer

import (
	"context"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestTransformManifests_ExposedNoOverrideUrl(t *testing.T) {
	tr, spinSvc, _ := th.setupTransformer(&exposeLbTransformerGenerator{}, t)
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

	err = tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", expected, t)
	expected.Spec.Ports[0].Port = int32(443)
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedAggregatedAnnotations(t *testing.T) {
	tr, spinSvc, _ := th.setupTransformer(&exposeLbTransformerGenerator{}, t)
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

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", expected, t)
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedServiceTypeOverridden(t *testing.T) {
	tr, spinSvc, _ := th.setupTransformer(&exposeLbTransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = "service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"
	spinSvc.Spec.Expose.Service.Overrides["gate"] = spinnakerv1alpha1.ExposeConfigServiceOverrides{
		Type: "NodePort",
	}

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", expected, t)
	expected.Spec.Type = "NodePort"
	expected.Annotations = nil
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_NotExposed(t *testing.T) {
	tr, spinSvc, _ := th.setupTransformer(&exposeLbTransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = ""

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", expected, t)
	expected.Annotations = nil
	expected.Spec.Type = "ClusterIP"
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedPortFromConfig(t *testing.T) {
	tr, spinSvc, _ := th.setupTransformer(&exposeLbTransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = "service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"
	spinSvc.Spec.Expose.Service.Port = 7777

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", expected, t)
	expected.Annotations = nil
	expected.Spec.Ports[0].Port = 7777
	expected.Spec.Ports[0].TargetPort = intstr.IntOrString{IntVal: 8084}
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedPortFromOverrides(t *testing.T) {
	tr, spinSvc, _ := th.setupTransformer(&exposeLbTransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = "service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"
	spinSvc.Spec.Expose.Service.Port = 7777
	spinSvc.Spec.Expose.Service.Overrides["gate"] = spinnakerv1alpha1.ExposeConfigServiceOverrides{Port: 1111}

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", expected, t)
	expected.Annotations = nil
	expected.Spec.Ports[0].Port = 1111
	expected.Spec.Ports[0].TargetPort = intstr.IntOrString{IntVal: 8084}
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

// Input: existing services running on default port, then spin config changes to custom port
func TestTransformHalconfig_ExposedPortAddedToConfig(t *testing.T) {
	gateSvc := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", gateSvc, t)
	gateSvc.Status.LoadBalancer.Ingress = append(gateSvc.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{Hostname: "abc.com"})
	fakeClient := fake.NewFakeClient(gateSvc)
	tr, spinSvc, hc := th.setupTransformerWithFakeClient(&exposeLbTransformerGenerator{}, fakeClient, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = "service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"
	spinSvc.Spec.Expose.Service.Port = 7777

	err := tr.TransformConfig(context.TODO())
	assert.Nil(t, err)

	actualHcUrl, err := hc.GetHalConfigPropString(context.TODO(), util.GateOverrideBaseUrlProp)
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com:7777", actualHcUrl)
	assert.Equal(t, "http://abc.com:7777", spinSvc.Status.APIUrl)
}

// Input: existing services running on custom port, then spin config changes the port
func TestTransformHalconfig_ExposedPortChanges(t *testing.T) {
	gateSvc := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", gateSvc, t)
	gateSvc.Spec.Ports[0].Port = 1111
	gateSvc.Status.LoadBalancer.Ingress = append(gateSvc.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{Hostname: "abc.com"})
	fakeClient := fake.NewFakeClient(gateSvc)
	tr, spinSvc, hc := th.setupTransformerWithFakeClient(&exposeLbTransformerGenerator{}, fakeClient, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = "service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"
	spinSvc.Spec.Expose.Service.Port = 7777

	err := tr.TransformConfig(context.TODO())
	assert.Nil(t, err)

	actualHcUrl, err := hc.GetHalConfigPropString(context.TODO(), util.GateOverrideBaseUrlProp)
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com:7777", actualHcUrl)
	assert.Equal(t, "http://abc.com:7777", spinSvc.Status.APIUrl)
}

// Input: existing services running on custom port, then spin config removes the custom port
func TestTransformHalconfig_ExposedPortRemovedFromConfig(t *testing.T) {
	gateSvc := &corev1.Service{}
	th.objectFromJson("output_service_lb.json", gateSvc, t)
	gateSvc.Spec.Ports[0].Port = 1111
	gateSvc.Status.LoadBalancer.Ingress = append(gateSvc.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{Hostname: "abc.com"})
	fakeClient := fake.NewFakeClient(gateSvc)
	tr, spinSvc, hc := th.setupTransformerWithFakeClient(&exposeLbTransformerGenerator{}, fakeClient, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addServiceToGenConfig(gen, "gate", "input_service.json", t)
	spinSvc.Spec.Expose.Type = "service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"
	spinSvc.Spec.Expose.Service.Port = 0

	err := tr.TransformConfig(context.TODO())
	assert.Nil(t, err)

	actualHcUrl, err := hc.GetHalConfigPropString(context.TODO(), util.GateOverrideBaseUrlProp)
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com", actualHcUrl)
	assert.Equal(t, "http://abc.com", spinSvc.Status.APIUrl)
}
