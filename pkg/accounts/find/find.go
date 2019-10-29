package find

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FindSpinnakerService(c client.Client, ns string, builder v1alpha2.SpinnakerServiceBuilderInterface) (v1alpha2.SpinnakerServiceInterface, error) {
	l := builder.NewList()
	if err := c.List(context.TODO(), l, client.InNamespace(ns)); err != nil {
		return nil, err
	}
	items := l.GetItems()
	if len(items) > 0 {
		return items[0], nil
	}
	return nil, nil
}

func FindDeployment(c client.Client, spinsvc v1alpha2.SpinnakerServiceInterface, service string) (*v12.Deployment, error) {
	dep := &v12.Deployment{}
	err := c.Get(context.TODO(), client.ObjectKey{Namespace: spinsvc.GetNamespace(), Name: fmt.Sprintf("spin-%s", service)}, dep)
	return dep, err
}

func FindSecretInDeployment(c client.Client, dep *v12.Deployment, containerName, path string) (*v1.Secret, error) {
	name := getMountedSecretNameInDeployment(dep, containerName, path)
	if name != "" {
		sec := &v1.Secret{}
		err := c.Get(context.TODO(), client.ObjectKey{Namespace: dep.Namespace, Name: name}, sec)
		return sec, err
	}
	return nil, fmt.Errorf("unable to find secret at path %s in container %s in deployment %s", path, containerName, dep.Name)
}

func getMountedSecretNameInDeployment(dep *v12.Deployment, containerName, path string) string {
	for _, c := range dep.Spec.Template.Spec.Containers {
		if c.Name == containerName {
			// Look for the volume mount here
			for _, vm := range c.VolumeMounts {
				if vm.MountPath == path {
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
			}
		}
	}
	return ""
}
