package changedetector

import (
	"encoding/json"
	"fmt"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"path/filepath"
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
	spinSvc := &spinnakerv1alpha1.SpinnakerService{}
	h.objectFromJson("spinsvc.json", spinSvc, t)
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
	spinSvc.Status.HalConfig = f
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

func (h *testHelpers) buildSvc(name string, svcType string, port int32) *corev1.Service {
	myAnnotations := map[string]string{
		"service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http",
	}
	portName := fmt.Sprintf("%s-tcp", name[len("spin-"):])
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "ns1",
			Annotations: myAnnotations,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceType(svcType),
			Ports: []corev1.ServicePort{
				{Name: portName, Port: port, TargetPort: intstr.IntOrString{IntVal: port}},
			},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{
				{Hostname: "acme.com"},
			}},
		},
	}
}

func (h *testHelpers) objectFromJson(fileName string, target interface{}, t *testing.T) {
	fileContents := h.loadJsonFile(fileName, t)
	err := json.Unmarshal([]byte(fileContents), target)
	if err != nil {
		t.Fatal(err)
	}
}

func (h *testHelpers) loadJsonFile(fileName string, t *testing.T) string {
	path := filepath.Join("testdata", fileName) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes)
}
