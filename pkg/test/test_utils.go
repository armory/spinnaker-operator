package test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/generated"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

func init() {
	v1alpha2.RegisterTypes()
}

var TypesFactory = interfaces.DefaultTypesFactory

type DummyK8sSecretEngine struct {
	Secret string
	File   bool
}

func (s *DummyK8sSecretEngine) Decrypt() (string, error) {
	return s.Secret, nil
}

func (s *DummyK8sSecretEngine) IsFile() bool {
	return s.File
}

func ManifestToSpinService(s string, t *testing.T) interfaces.SpinnakerService {
	svc := TypesFactory.NewService()
	ReadYamlString([]byte(s), svc, t)
	return svc
}

func ManifestFileToSpinService(manifestYaml string, t *testing.T) interfaces.SpinnakerService {
	svc := TypesFactory.NewService()
	ReadYamlFile(manifestYaml, svc, t)
	return svc
}

func ReadYamlString(s []byte, target interface{}, t *testing.T) {
	err := yaml.Unmarshal(s, target)
	if err != nil {
		t.Fatal(err)
	}
}

func ReadYamlFile(path string, target interface{}, t *testing.T) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ReadYamlString(bytes, target, t)
}

func FakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	return fake.NewFakeClientWithScheme(scheme.Scheme, objs...)
}

func BuildSvc(name string, svcType string, publicPort int32, t *testing.T) *corev1.Service {
	svc := &corev1.Service{}
	ReadYamlFile("testdata/service.yml", svc, t)
	svc.Name = name
	svc.Spec.Selector["cluster"] = name
	svc.Spec.Type = corev1.ServiceType(svcType)
	svc.Spec.Ports[0].Port = publicPort
	portName := fmt.Sprintf("%s-tcp", name[len("spin-"):])
	svc.Spec.Ports[0].Name = portName
	return svc
}

func AddDeploymentToGenConfig(gen *generated.SpinnakerGeneratedConfig, depName string, fileName string, t *testing.T) {
	if gen.Config == nil {
		gen.Config = make(map[string]generated.ServiceConfig)
	}
	dep := &v1.Deployment{}
	ReadYamlFile(fileName, dep, t)
	gen.Config[depName] = generated.ServiceConfig{
		Deployment: dep,
	}
}

func AddServiceToGenConfig(gen *generated.SpinnakerGeneratedConfig, svcName string, fileName string, t *testing.T) {
	if gen.Config == nil {
		gen.Config = make(map[string]generated.ServiceConfig)
	}
	svc := &corev1.Service{}
	ReadYamlFile(fileName, svc, t)
	gen.Config[svcName] = generated.ServiceConfig{
		Service: svc,
	}
}

func GetSpinnakerService() (interfaces.SpinnakerService, error) {
	s := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
 name: test
spec:
 spinnakerConfig:
   config:
     providers:
       enabled: true
       dockerRegistry:
         accounts:
         - name: dockerhub
           requiredGroupMembership: []
           providerVersion: V1
           permissions: {}
           address: https://index.docker.io
           username: user
           email: test@spinnaker.io
           cacheIntervalSeconds: 120
           clientTimeoutMillis: 120000
           cacheThreads: 2
           paginateSize: 100
           sortTagsByDate: true
           trackDigests: true
           insecureRegistry: false
           repositories:
             - org/image-1
             - org/image-2
`
	spinsvc := interfaces.DefaultTypesFactory.NewService()
	err := yaml.Unmarshal([]byte(s), spinsvc)
	if err != nil {
		return nil, err
	}
	return spinsvc, nil
}
