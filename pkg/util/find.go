package util

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory/go-yaml-tools/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/ghodss/yaml"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errSecretNotFound = errors.New("secret not found")

func FindSpinnakerService(c client.Client, ns string, builder interfaces.TypesFactory) (interfaces.SpinnakerService, error) {
	l := builder.NewServiceList()
	if err := c.List(context.TODO(), l, client.InNamespace(ns)); err != nil {
		return nil, err
	}
	items := l.GetItems()
	if len(items) > 0 {
		return items[0], nil
	}
	return nil, nil
}

func FindDeployment(c client.Client, spinsvc interfaces.SpinnakerService, service string) (*v12.Deployment, error) {
	dep := &v12.Deployment{}
	err := c.Get(context.TODO(), client.ObjectKey{Namespace: spinsvc.GetNamespace(), Name: fmt.Sprintf("spin-%s", service)}, dep)
	return dep, err
}

func FindSecretInDeployment(c client.Client, dep *v12.Deployment, containerName, path string) (*v1.Secret, error) {
	name := GetMountedSecretNameInDeployment(dep, containerName, path)
	if name != "" {
		sec := &v1.Secret{}
		err := c.Get(context.TODO(), client.ObjectKey{Namespace: dep.Namespace, Name: name}, sec)
		return sec, err
	}
	return nil, fmt.Errorf("unable to find secret at path %s in container %s in deployment %s", path, containerName, dep.Name)
}

func GetSecretContent(c *rest.Config, namespace, name, key string) (string, error) {
	cl, err := clientcorev1.NewForConfig(c)
	if err != nil {
		return "", err
	}
	sec, err := cl.Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if d, ok := sec.Data[key]; ok {
		return string(d), nil
	}
	return "", errSecretNotFound
}

func GetMountedSecretNameInDeployment(dep *v12.Deployment, containerName, path string) string {
	container := GetContainerInDeployment(dep, containerName)
	if container == nil {
		return ""
	}
	// Look for the volume mount here
	for _, vm := range container.VolumeMounts {
		if vm.MountPath != path {
			continue
		}
		// Look for the secret
		for _, v := range dep.Spec.Template.Spec.Volumes {
			if v.Name == vm.Name {
				if v.Secret != nil {
					return v.Secret.SecretName
				}
				return ""
			}
		}
	}
	return ""
}

func GetContainerInDeployment(dep *v12.Deployment, containerName string) *v1.Container {
	for i := range dep.Spec.Template.Spec.Containers {
		c := &dep.Spec.Template.Spec.Containers[i]
		if c.Name == containerName {
			return c
		}
	}
	return nil
}

func UpdateSecret(secret *v1.Secret, settings map[string]interface{}, fileName string) error {
	b, err := yaml.Marshal(settings)
	if err != nil {
		return err
	}
	secret.Data[fileName] = b
	return nil
}

// GetServiceAccountData returns the service account token and temp path to root ca
func GetServiceAccountData(ctx context.Context, name, ns string, c client.Client) (string, string, error) {
	list := &v1.SecretList{}
	opts := client.InNamespace(ns)
	if err := c.List(ctx, list, opts); err != nil {
		return "", "", err
	}
	for i := 0; i < len(list.Items); i++ {
		s := list.Items[i]
		if s.Type != v1.SecretTypeServiceAccountToken {
			continue
		}
		saName := s.Annotations[v1.ServiceAccountNameKey]
		if saName != name {
			continue
		}
		token := string(s.Data[v1.ServiceAccountTokenKey])
		caBytes := s.Data[v1.ServiceAccountRootCAKey]
		caPath, err := secrets.ToTempFile(caBytes)
		if err != nil {
			return "", "", err
		}
		return token, caPath, nil
	}
	return "", "", fmt.Errorf("no secret for service account %s was found on namespace %s", name, ns)
}

func GetSpinnakerServices(list interfaces.SpinnakerServiceList, ns string, c client.Client) ([]interfaces.SpinnakerService, error) {
	var opts client.ListOption
	opts = client.InNamespace(ns)
	err := c.List(context.TODO(), list, opts)
	if err != nil {
		return nil, err
	}
	return list.GetItems(), nil
}

func GetSecretForDefaultConfigPath(config generated.ServiceConfig, container string) *v1.Secret {
	return GetSecretForPath(config, container, "/opt/spinnaker/config")
}

func GetSecretForPath(config generated.ServiceConfig, container, path string) *v1.Secret {
	secretName := GetMountedSecretNameInDeployment(config.Deployment, container, path)
	if secretName != "" {
		for i := range config.Resources {
			o := config.Resources[i]
			if sc, ok := o.(*v1.Secret); ok && sc.GetName() == secretName {
				return sc
			}
		}
	}
	return nil
}
