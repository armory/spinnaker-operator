package util

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestAddEnvVarToDeployment(t *testing.T) {
	cases := []struct {
		name   string
		dep    *appsv1.Deployment
		e      v1.EnvVar
		merge  func(old, new string) string
		filter func(c v1.Container) bool
		assert func(d *appsv1.Deployment, t *testing.T)
	}{
		{
			name:   "append to existing var",
			dep:    newDeployment(),
			e:      v1.EnvVar{Name: "ENV_VAR_EXISTING", Value: " addedValue"},
			merge:  func(old, new string) string { return fmt.Sprintf("%s%s", old, new) },
			filter: func(c v1.Container) bool { return c.Name == "first-container" },
			assert: func(d *appsv1.Deployment, t *testing.T) {
				ctrs := d.Spec.Template.Spec.Containers
				assert.Equal(t, 1, len(ctrs[0].Env))
				assert.Equal(t, "ENV_VAR_EXISTING", ctrs[0].Env[0].Name)
				assert.Equal(t, "existingValue addedValue", ctrs[0].Env[0].Value)
				assert.Equal(t, 0, len(ctrs[1].Env))
			},
		},
		{
			name:   "create new var",
			dep:    newDeployment(),
			merge:  func(old, new string) string { return new },
			e:      v1.EnvVar{Name: "NEW_VAR", Value: "newValue"},
			filter: func(c v1.Container) bool { return c.Name == "first-container" },
			assert: func(d *appsv1.Deployment, t *testing.T) {
				ctrs := d.Spec.Template.Spec.Containers
				assert.Equal(t, 2, len(ctrs[0].Env))
				assert.Equal(t, "ENV_VAR_EXISTING", ctrs[0].Env[0].Name)
				assert.Equal(t, "existingValue", ctrs[0].Env[0].Value)
				assert.Equal(t, "NEW_VAR", ctrs[0].Env[1].Name)
				assert.Equal(t, "newValue", ctrs[0].Env[1].Value)
				assert.Equal(t, 0, len(ctrs[1].Env))
			},
		},
		{
			name:   "overwrite var",
			dep:    newDeployment(),
			merge:  func(old, new string) string { return new },
			e:      v1.EnvVar{Name: "ENV_VAR_EXISTING", Value: "newValue"},
			filter: func(c v1.Container) bool { return c.Name == "first-container" },
			assert: func(d *appsv1.Deployment, t *testing.T) {
				ctrs := d.Spec.Template.Spec.Containers
				assert.Equal(t, 1, len(ctrs[0].Env))
				assert.Equal(t, "ENV_VAR_EXISTING", ctrs[0].Env[0].Name)
				assert.Equal(t, "newValue", ctrs[0].Env[0].Value)
				assert.Equal(t, 0, len(ctrs[1].Env))
			},
		},
		{
			name:   "add to all containers",
			dep:    newDeployment(),
			merge:  func(old, new string) string { return new },
			e:      v1.EnvVar{Name: "NEW_VAR", Value: "newValue"},
			filter: func(c v1.Container) bool { return true },
			assert: func(d *appsv1.Deployment, t *testing.T) {
				ctrs := d.Spec.Template.Spec.Containers
				assert.Equal(t, 2, len(ctrs[0].Env))
				assert.Equal(t, "ENV_VAR_EXISTING", ctrs[0].Env[0].Name)
				assert.Equal(t, "existingValue", ctrs[0].Env[0].Value)
				assert.Equal(t, "NEW_VAR", ctrs[0].Env[1].Name)
				assert.Equal(t, "newValue", ctrs[0].Env[1].Value)
				assert.Equal(t, 1, len(ctrs[1].Env))
				assert.Equal(t, "NEW_VAR", ctrs[1].Env[0].Name)
				assert.Equal(t, "newValue", ctrs[1].Env[0].Value)
			},
		},
		{
			name:   "idempotent call",
			dep:    newDeployment(),
			merge:  func(old, new string) string { return new },
			e:      v1.EnvVar{Name: "NEW_VAR", Value: "newValue"},
			filter: func(c v1.Container) bool { return c.Name == "first-container" },
			assert: func(d *appsv1.Deployment, t *testing.T) {
				AddEnvVarToDeployment(d, v1.EnvVar{Name: "NEW_VAR", Value: "newValue"},
					func(old, new string) string { return new },
					func(c v1.Container) bool { return c.Name == "first-container" })
				ctrs := d.Spec.Template.Spec.Containers
				assert.Equal(t, 2, len(ctrs[0].Env))
				assert.Equal(t, "ENV_VAR_EXISTING", ctrs[0].Env[0].Name)
				assert.Equal(t, "existingValue", ctrs[0].Env[0].Value)
				assert.Equal(t, "NEW_VAR", ctrs[0].Env[1].Name)
				assert.Equal(t, "newValue", ctrs[0].Env[1].Value)
				assert.Equal(t, 0, len(ctrs[1].Env))
			},
		},
	}

	for _, c := range cases {
		AddEnvVarToDeployment(c.dep, c.e, c.merge, c.filter)
		c.assert(c.dep, t)
	}
}

func newDeployment() *appsv1.Deployment {
	return newDeploymentWithMountPath("/opt/config")
}

func newDeploymentWithMountPath(mountPath string) *appsv1.Deployment {
	return &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:         "first-container",
							Env:          []v1.EnvVar{{Name: "ENV_VAR_EXISTING", Value: "existingValue"}},
							VolumeMounts: []v1.VolumeMount{{Name: "first-volume", MountPath: mountPath}},
						},
						{
							Name: "second-container",
						},
					},
					Volumes: []v1.Volume{{
						Name:         "first-volume",
						VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "first-volume"}},
					}}},
			},
		},
	}
}
