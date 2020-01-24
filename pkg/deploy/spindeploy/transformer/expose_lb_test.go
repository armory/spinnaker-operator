package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestTransformManifests_ExposedNoOverrideUrl(t *testing.T) {
	tr, _ := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_expose.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", expected, t)
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedWithOverrideUrlChangingPort(t *testing.T) {
	tr, spinSvc := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_expose.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)
	err := spinSvc.Spec.SpinnakerConfig.SetHalConfigProp("security.apiSecurity.overrideBaseUrl", "https://my-api.spin.com")

	err = tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", expected, t)
	expected.Spec.Ports[0].Port = int32(443)
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedAggregatedAnnotations(t *testing.T) {
	tr, spinSvc := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_expose.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)
	spinSvc.Spec.Expose.Service.Overrides["gate"] = v1alpha2.ExposeConfigServiceOverrides{
		Annotations: map[string]string{
			"service.beta.kubernetes.io/aws-load-balancer-ssl-ports": "80,443",
		},
	}

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", expected, t)
	expected.Annotations = map[string]string{
		"service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http",
		"service.beta.kubernetes.io/aws-load-balancer-ssl-ports":        "80,443",
	}
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedAnnotationsRemoved(t *testing.T) {
	tr, spinSvc := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_expose.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/output_service_lb.yml", t)
	spinSvc.Spec.Expose.Service.Annotations = map[string]string{} // annotations removed from config

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", expected, t)
	expected.Annotations = map[string]string{}
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedServiceTypeOverridden(t *testing.T) {
	tr, spinSvc := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_expose.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)
	spinSvc.Spec.Expose.Service.Overrides["gate"] = v1alpha2.ExposeConfigServiceOverrides{
		Type: "NodePort",
	}

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", expected, t)
	expected.Spec.Type = "NodePort"
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_NotExposed(t *testing.T) {
	tr, _ := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_minimal.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", expected, t)
	expected.Annotations = nil
	expected.Spec.Type = "ClusterIP"
	expected.Spec.Ports[0].Port = 8084
	expected.Spec.Ports[0].Name = ""
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedPortFromConfig(t *testing.T) {
	tr, spinSvc := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_expose.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)
	spinSvc.Spec.Expose.Service.PublicPort = 7777

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", expected, t)
	expected.Spec.Ports[0].Port = 7777
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

func TestTransformManifests_ExposedPortFromOverrides(t *testing.T) {
	tr, spinSvc := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_expose.yml", t)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)
	spinSvc.Spec.Expose.Service.PublicPort = 7777
	spinSvc.Spec.Expose.Service.Overrides["gate"] = v1alpha2.ExposeConfigServiceOverrides{PublicPort: 1111}

	err := tr.TransformManifests(context.TODO(), nil, gen)
	assert.Nil(t, err)

	expected := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", expected, t)
	expected.Spec.Ports[0].Port = 1111
	assert.Equal(t, expected, gen.Config["gate"].Service)
}

// Input: existing services running on default port, then spin config changes to custom port
func TestTransformHalconfig_ExposedPortAddedToConfig(t *testing.T) {
	gateSvc := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", gateSvc, t)
	gateSvc.Status.LoadBalancer.Ingress = append(gateSvc.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{Hostname: "abc.com"})
	tr, spinSvc := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_expose.yml", t, gateSvc)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)
	spinSvc.Spec.Expose.Service.PublicPort = 7777

	err := tr.TransformConfig(context.TODO())
	assert.Nil(t, err)

	actualOverrideUrl, err := (tr.(*exposeLbTransformer)).svc.GetSpinnakerConfig().GetHalConfigPropString(context.TODO(), util.GateOverrideBaseUrlProp)
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com:7777", actualOverrideUrl)
	assert.Equal(t, "http://abc.com:7777", (tr.(*exposeLbTransformer)).svc.GetStatus().APIUrl)
}

// Input: existing services running on default port, then spin config changes to custom port on override section
func TestTransformHalconfig_ExposedPortOverrideAddedToConfig(t *testing.T) {
	gateSvc := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", gateSvc, t)
	gateSvc.Status.LoadBalancer.Ingress = append(gateSvc.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{Hostname: "abc.com"})
	tr, spinSvc := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_expose.yml", t, gateSvc)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)
	spinSvc.Spec.Expose.Service.Overrides["gate"] = v1alpha2.ExposeConfigServiceOverrides{PublicPort: 7777}

	err := tr.TransformConfig(context.TODO())
	assert.Nil(t, err)

	actualOverrideUrl, err := (tr.(*exposeLbTransformer)).svc.GetSpinnakerConfig().GetHalConfigPropString(context.TODO(), util.GateOverrideBaseUrlProp)
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com:7777", actualOverrideUrl)
	assert.Equal(t, "http://abc.com:7777", (tr.(*exposeLbTransformer)).svc.GetStatus().APIUrl)
}

// Input: existing services running on custom port, then spin config changes the port
func TestTransformHalconfig_ExposedPortChanges(t *testing.T) {
	gateSvc := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", gateSvc, t)
	gateSvc.Spec.Ports[0].Port = 1111
	gateSvc.Status.LoadBalancer.Ingress = append(gateSvc.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{Hostname: "abc.com"})
	tr, spinSvc := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_expose.yml", t, gateSvc)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)
	spinSvc.Spec.Expose.Service.PublicPort = 7777

	err := tr.TransformConfig(context.TODO())
	assert.Nil(t, err)

	actualOverrideUrl, err := (tr.(*exposeLbTransformer)).svc.GetSpinnakerConfig().GetHalConfigPropString(context.TODO(), util.GateOverrideBaseUrlProp)
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com:7777", actualOverrideUrl)
	assert.Equal(t, "http://abc.com:7777", (tr.(*exposeLbTransformer)).svc.GetStatus().APIUrl)
}

// Input: existing services running on custom port, then spin config removes the custom port
func TestTransformHalconfig_ExposedPortRemovedFromConfig(t *testing.T) {
	gateSvc := &corev1.Service{}
	test.ReadYamlFile("testdata/output_service_lb.yml", gateSvc, t)
	gateSvc.Spec.Ports[0].Port = 1111
	gateSvc.Status.LoadBalancer.Ingress = append(gateSvc.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{Hostname: "abc.com"})
	tr, spinSvc := th.setupTransformer(&exposeLbTransformerGenerator{}, "testdata/spinsvc_expose.yml", t, gateSvc)
	gen := &generated.SpinnakerGeneratedConfig{}
	test.AddServiceToGenConfig(gen, "gate", "testdata/input_service.yml", t)
	spinSvc.Spec.Expose.Service.PublicPort = 0

	err := tr.TransformConfig(context.TODO())
	assert.Nil(t, err)

	actualOverrideUrl, err := (tr.(*exposeLbTransformer)).svc.GetSpinnakerConfig().GetHalConfigPropString(context.TODO(), util.GateOverrideBaseUrlProp)
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com", actualOverrideUrl)
	assert.Equal(t, "http://abc.com", (tr.(*exposeLbTransformer)).svc.GetStatus().APIUrl)
}
