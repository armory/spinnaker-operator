package test

import (
	"encoding/json"
	"fmt"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func ReadYamlFile(path string, t *testing.T) interface{} {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var hc interface{}
	err = yaml.Unmarshal(bytes, &hc)
	if err != nil {
		t.Fatal(err)
	}
	return hc
}

func SetupSpinnakerConfig(halConfigFile string, t *testing.T) *halconfig.SpinnakerConfig {
	hc := ReadYamlFile(halConfigFile, t)
	config := halconfig.SpinnakerConfig{
		HalConfig: hc,
		Profiles:  map[string]interface{}{},
	}
	return &config
}

func AddProfileToConfig(profileName string, profileFile string, spinConfig *halconfig.SpinnakerConfig, t *testing.T) {
	profile := ReadYamlFile(profileFile, t)
	spinConfig.Profiles[profileName] = profile
}

func SetupSpinnakerService(serviceJsonFile string, halConfigFile string, t *testing.T) (*spinnakerv1alpha1.SpinnakerService, *halconfig.SpinnakerConfig, *corev1.ConfigMap) {
	spinSvc := &spinnakerv1alpha1.SpinnakerService{}
	ObjectFromJson(serviceJsonFile, spinSvc, t)
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
	config := SetupSpinnakerConfig(halConfigFile, t)
	return spinSvc, config, cm
}

func ObjectFromJson(fileName string, target interface{}, t *testing.T) {
	fileContents := LoadJsonFile(fileName, t)
	err := json.Unmarshal([]byte(fileContents), target)
	if err != nil {
		t.Fatal(err)
	}
}

func LoadJsonFile(fileName string, t *testing.T) string {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes)
}

func BuildSvc(name string, svcType string, port int32) *corev1.Service {
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

func AddDeploymentToGenConfig(gen *generated.SpinnakerGeneratedConfig, depName string, fileName string, t *testing.T) {
	if gen.Config == nil {
		gen.Config = make(map[string]generated.ServiceConfig)
	}
	dep := &v1beta2.Deployment{}
	ObjectFromJson(fileName, dep, t)
	gen.Config[depName] = generated.ServiceConfig{
		Deployment: dep,
	}
}

func AddServiceToGenConfig(gen *generated.SpinnakerGeneratedConfig, svcName string, fileName string, t *testing.T) {
	if gen.Config == nil {
		gen.Config = make(map[string]generated.ServiceConfig)
	}
	svc := &corev1.Service{}
	ObjectFromJson(fileName, svc, t)
	gen.Config[svcName] = generated.ServiceConfig{
		Service: svc,
	}
}
