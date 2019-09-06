package changedetector

import (
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

type testHelpers struct{}

var th = testHelpers{}

func (h *testHelpers) setupChangeDetector(generator Generator, client client.Client, t *testing.T) ChangeDetector {
	ch, _ := generator.NewChangeDetector(client, log.Log.WithName("spinnakerservice"))
	return ch
}

func (h *testHelpers) buildSpinSvc(t *testing.T) (*spinnakerv1alpha1.SpinnakerService, *corev1.ConfigMap, *halconfig.SpinnakerConfig) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "myconfig",
			Namespace:       "ns1",
			ResourceVersion: "123456",
		},
	}
	f := spinnakerv1alpha1.SpinnakerFileSourceStatus{
		ConfigMap: &spinnakerv1alpha1.SpinnakerFileSourceReferenceStatus{
			Name:            "myconfig",
			Namespace:       "ns1",
			ResourceVersion: "123456",
		},
	}
	spinSvc := &spinnakerv1alpha1.SpinnakerService{
		Status: spinnakerv1alpha1.SpinnakerServiceStatus{HalConfig: f},
	}
	config := h.setupSpinnakerConfig(t)
	return spinSvc, cm, config
}

func (h *testHelpers) setupSpinnakerConfig(t *testing.T) *halconfig.SpinnakerConfig {
	path := "testdata/halconfig.yml"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var hc interface{}
	err = yaml.Unmarshal(bytes, &hc)
	if err != nil {
		t.Fatal(err)
	}

	path = "testdata/profile_gate.yml"
	bytes, err = ioutil.ReadFile(path)
	var profile interface{}
	err = yaml.Unmarshal(bytes, &profile)
	if err != nil {
		t.Fatal(err)
	}
	config := halconfig.SpinnakerConfig{
		HalConfig: hc,
		Profiles:  map[string]interface{}{},
	}
	config.Profiles["gate"] = profile
	return &config
}

func (h *testHelpers) buildSvc(name string, svcType string, annotations map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "ns1",
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceType(svcType),
			Ports: []corev1.ServicePort{
				{Name: name + "-tcp", Port: 80},
			},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{
				{Hostname: "acme.com"},
			}},
		},
	}
}
