package util

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

// AddEnvVarToDeployment adds an environment variable to the given deployment containers for which the filter function returns true
func AddEnvVarToDeployment(d *appsv1.Deployment, e v1.EnvVar, appendIfExists bool, filter func(c v1.Container) bool) {
	ctrs := make([]v1.Container, 0)
	for _, c := range d.Spec.Template.Spec.Containers {
		if !filter(c) {
			ctrs = append(ctrs, c)
			continue
		}
		found := false
		vars := make([]v1.EnvVar, 0)
		for _, ce := range c.Env {
			if ce.Name == e.Name {
				found = true
				if appendIfExists {
					ce.Value = fmt.Sprintf("%s%s", ce.Value, e.Value)
				} else {
					ce.Value = e.Value
				}
			}
			vars = append(vars, ce)
		}
		if !found {
			vars = append(vars, e)
		}
		c.Env = vars
		ctrs = append(ctrs, c)
	}
	d.Spec.Template.Spec.Containers = ctrs
}

// AddVolumeMountAsSecret creates a new volume and volume mount for the given secret and adds it to all containers for which the filter function returns true
func AddVolumeMountAsSecret(d *appsv1.Deployment, s *v1.Secret, mountPath string, filter func(c v1.Container) bool) {
	addSecretVolumeIfNeeded(s.Name, d)
	addSecretMountIfNeeded(s.Name, mountPath, d, filter)
}

func addSecretVolumeIfNeeded(secretName string, d *appsv1.Deployment) {
	vols := make([]v1.Volume, 0)
	for _, vi := range d.Spec.Template.Spec.Volumes {
		vols = append(vols, vi)
		if vi.Name == secretName {
			return
		}
	}
	vols = append(vols, v1.Volume{
		Name: secretName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	})
	d.Spec.Template.Spec.Volumes = vols
}

func addSecretMountIfNeeded(volumeName, mountPath string, d *appsv1.Deployment, filter func(c v1.Container) bool) {
	ctrs := make([]v1.Container, 0)
	for _, c := range d.Spec.Template.Spec.Containers {
		if !filter(c) {
			ctrs = append(ctrs, c)
			continue
		}
		added := false
		for _, vm := range c.VolumeMounts {
			if vm.Name == volumeName && vm.MountPath == mountPath {
				added = true
			}
		}
		if !added {
			c.VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{Name: volumeName, MountPath: mountPath})
		}
		ctrs = append(ctrs, c)
	}
	d.Spec.Template.Spec.Containers = ctrs
}
