package util

import (
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
	}

	for _, c := range cases {
		AddEnvVarToDeployment(c.dep, c.e, c.append, c.filter)
		c.assert(c.dep, t)
	}
}

func newDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "first-container",
							Env:  []v1.EnvVar{{Name: "ENV_VAR_EXISTING", Value: "existingValue"}},
						},
						{
							Name: "second-container",
						},
					},
				},
			},
		},
	}
}
