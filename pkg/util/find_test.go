package util

import (
	"github.com/magiconair/properties/assert"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestGetMountedSecretNameInDeployment(t *testing.T) {
	d := &v1.Deployment{
		Spec: v1.DeploymentSpec{
			Replicas: nil,
			Selector: nil,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "test1",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "val1",
									Items:      nil,
								},
							},
						},
						{
							Name: "test2",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "val2",
									Items:      nil,
								},
							},
						},
						{
							Name: "test3",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "val3",
									Items:      nil,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name: "monitoring",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "test1",
									MountPath: "/opt/spinnaker/config",
								},
								{
									Name:      "test2",
									MountPath: "/opt/monitoring",
								},
							},
						},
						{
							Name: "clouddriver",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "test3",
									MountPath: "/opt/spinnaker/config",
								},
								{
									Name:      "test1",
									MountPath: "/opt/monitoring",
								},
							},
						},
					},
				},
			},
		},
	}

	v := GetMountedSecretNameInDeployment(d, "clouddriver", "/opt/spinnaker/config")
	assert.Equal(t, "val3", v)
}
