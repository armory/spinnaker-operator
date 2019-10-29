package changedetector

import (
	"encoding/json"
	"fmt"
	spinnakerv1alpha2 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
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

func (h *testHelpers) buildSpinSvc(t *testing.T) *spinnakerv1alpha2.SpinnakerService {
	ss := &spinnakerv1alpha2.SpinnakerService{}
	h.objectFromJson("spinsvc.json", ss, t)
	return ss
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
	fileContents := h.loadFileContent(fileName, t)
	err := json.Unmarshal([]byte(fileContents), target)
	if err != nil {
		t.Fatal(err)
	}
}

func (h *testHelpers) loadFileContent(fileName string, t *testing.T) string {
	path := filepath.Join("testdata", fileName) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes)
}
