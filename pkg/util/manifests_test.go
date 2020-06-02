package util

import (
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestAddEnvVarToDeployment(t *testing.T) {
	cases := []struct {
		name   string
		dep    *appsv1.Deployment
		e      v1.EnvVar
		append bool
		filter func(c v1.Container) bool
		assert func(d *appsv1.Deployment, t *testing.T)
	}{
		{
			name:   "append to existing var",
			dep:    newDeployment(),
			append: true,
			e:      v1.EnvVar{Name: "ENV_VAR_EXISTING", Value: " addedValue"},
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
			append: false,
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
			append: false,
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
			append: false,
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
			append: false,
			e:      v1.EnvVar{Name: "NEW_VAR", Value: "newValue"},
			filter: func(c v1.Container) bool { return c.Name == "first-container" },
			assert: func(d *appsv1.Deployment, t *testing.T) {
				AddEnvVarToDeployment(d, v1.EnvVar{Name: "NEW_VAR", Value: "newValue"}, false, func(c v1.Container) bool { return c.Name == "first-container" })
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
		AddEnvVarToDeployment(c.dep, c.e, c.append, c.filter)
		c.assert(c.dep, t)
	}
}

func TestAddVolumeMountAsSecret(t *testing.T) {
	cases := []struct {
		name   string
		d      *appsv1.Deployment
		s      *v1.Secret
		filter func(c v1.Container) bool
		assert func(d *appsv1.Deployment, t *testing.T)
	}{
		{
			name:   "add non existing volume",
			d:      newDeployment(),
			s:      newSecret(),
			filter: func(c v1.Container) bool { return true },
			assert: func(d *appsv1.Deployment, t *testing.T) {
				volumes := d.Spec.Template.Spec.Volumes
				assert.Equal(t, 2, len(volumes))
				assert.Equal(t, "test-secret", volumes[1].Name)
				assert.Equal(t, "test-secret", volumes[1].Secret.SecretName)
				ctrs := d.Spec.Template.Spec.Containers
				assert.Equal(t, 2, len(ctrs[0].VolumeMounts))
				assert.Equal(t, "test-secret", ctrs[0].VolumeMounts[1].Name)
				assert.Equal(t, "/opt/extra", ctrs[0].VolumeMounts[1].MountPath)
				assert.Equal(t, 1, len(ctrs[1].VolumeMounts))
				assert.Equal(t, "test-secret", ctrs[1].VolumeMounts[0].Name)
				assert.Equal(t, "/opt/extra", ctrs[1].VolumeMounts[0].MountPath)
			},
		},
		{
			name:   "add volume to only one container",
			d:      newDeployment(),
			s:      newSecret(),
			filter: func(c v1.Container) bool { return c.Name == "first-container" },
			assert: func(d *appsv1.Deployment, t *testing.T) {
				volumes := d.Spec.Template.Spec.Volumes
				assert.Equal(t, 2, len(volumes))
				assert.Equal(t, "test-secret", volumes[1].Name)
				assert.Equal(t, "test-secret", volumes[1].Secret.SecretName)
				ctrs := d.Spec.Template.Spec.Containers
				assert.Equal(t, 2, len(ctrs[0].VolumeMounts))
				assert.Equal(t, "test-secret", ctrs[0].VolumeMounts[1].Name)
				assert.Equal(t, "/opt/extra", ctrs[0].VolumeMounts[1].MountPath)
				assert.Equal(t, 0, len(ctrs[1].VolumeMounts))
			},
		},
		{
			name:   "volume already added",
			d:      newDeploymentWithMountPath("/opt/extra"),
			s:      newSecretWithName("first-volume"),
			filter: func(c v1.Container) bool { return c.Name == "first-container" },
			assert: func(d *appsv1.Deployment, t *testing.T) {
				volumes := d.Spec.Template.Spec.Volumes
				assert.Equal(t, 1, len(volumes))
				assert.Equal(t, "first-volume", volumes[0].Name)
				ctrs := d.Spec.Template.Spec.Containers
				assert.Equal(t, 1, len(ctrs[0].VolumeMounts))
				assert.Equal(t, "first-volume", ctrs[0].VolumeMounts[0].Name)
				assert.Equal(t, "/opt/extra", ctrs[0].VolumeMounts[0].MountPath)
				assert.Equal(t, 0, len(ctrs[1].VolumeMounts))
			},
		},
		{
			name:   "idempotent call",
			d:      newDeployment(),
			s:      newSecret(),
			filter: func(c v1.Container) bool { return c.Name == "first-container" },
			assert: func(d *appsv1.Deployment, t *testing.T) {
				AddVolumeMountAsSecret(d, newSecret(), "/opt/extra", func(c v1.Container) bool { return c.Name == "first-container" })
				volumes := d.Spec.Template.Spec.Volumes
				assert.Equal(t, 2, len(volumes))
				assert.Equal(t, "test-secret", volumes[1].Name)
				assert.Equal(t, "test-secret", volumes[1].Secret.SecretName)
				ctrs := d.Spec.Template.Spec.Containers
				assert.Equal(t, 2, len(ctrs[0].VolumeMounts))
				assert.Equal(t, "test-secret", ctrs[0].VolumeMounts[1].Name)
				assert.Equal(t, "/opt/extra", ctrs[0].VolumeMounts[1].MountPath)
				assert.Equal(t, 0, len(ctrs[1].VolumeMounts))
			},
		},
	}

	for _, c := range cases {
		AddVolumeMountAsSecret(c.d, c.s, "/opt/extra", c.filter)
		c.assert(c.d, t)
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

func newSecret() *v1.Secret {
	return newSecretWithName("test-secret")
}

func newSecretWithName(name string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}
